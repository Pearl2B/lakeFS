package index

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/treeverse/lakefs/db"
	"github.com/treeverse/lakefs/ident"
	"github.com/treeverse/lakefs/index/dag"
	indexerrors "github.com/treeverse/lakefs/index/errors"
	"github.com/treeverse/lakefs/index/merkle"
	"github.com/treeverse/lakefs/index/model"
	pth "github.com/treeverse/lakefs/index/path"
	"github.com/treeverse/lakefs/index/store"
	"github.com/treeverse/lakefs/logging"
)

const (
	// DefaultPartialCommitRatio is the ratio (1/?) of writes that will trigger a partial commit (number between 0-1)
	DefaultPartialCommitRatio = 0.01

	// DefaultBranch is the branch to be automatically created when a repo is born
	DefaultBranch = "master"
)

type Index interface {
	WithContext(ctx context.Context) Index
	ReadObject(repoId, ref, path string, readUncommitted bool) (*model.Object, error)
	ReadEntryObject(repoId, ref, path string, readUncommitted bool) (*model.Entry, error)
	ReadEntryTree(repoId, ref, path string, readUncommitted bool) (*model.Entry, error)
	ReadRootObject(repoId, ref string, readUncommitted bool) (*model.Root, error)
	WriteObject(repoId, branch, path string, object *model.Object) error
	WriteEntry(repoId, branch, path string, entry *model.Entry) error
	WriteFile(repoId, branch, path string, entry *model.Entry, obj *model.Object) error
	DeleteObject(repoId, branch, path string) error
	ListObjectsByPrefix(repoId, ref, path, after string, results int, descend, readUncommitted bool) ([]*model.Entry, bool, error)
	ListBranchesByPrefix(repoId string, prefix string, amount int, after string) ([]*model.Branch, bool, error)
	ResetBranch(repoId, branch string) error
	CreateBranch(repoId, branch, ref string) (*model.Branch, error)
	GetBranch(repoId, branch string) (*model.Branch, error)
	Commit(repoId, branch, message, committer string, metadata map[string]string) (*model.Commit, error)
	GetCommit(repoId, commitId string) (*model.Commit, error)
	GetCommitLog(repoId, fromCommitId string, results int, after string) ([]*model.Commit, bool, error)
	DeleteBranch(repoId, branch string) error
	Diff(repoId, leftRef, rightRef string) (merkle.Differences, error)
	DiffWorkspace(repoId, branch string) (merkle.Differences, error)
	RevertCommit(repoId, branch, commit string) error
	RevertPath(repoId, branch, path string) error
	RevertObject(repoId, branch, path string) error
	Merge(repoId, source, destination, userId string) (merkle.Differences, error)
	CreateRepo(repoId, bucketName, defaultBranch string) error
	ListRepos(amount int, after string) ([]*model.Repo, bool, error)
	GetRepo(repoId string) (*model.Repo, error)
	DeleteRepo(repoId string) error
	CreateDedupEntryIfNone(repoId string, dedupId string, objName string) (string, error)
	CreateMultiPartUpload(repoId, id, path, objectName string, creationDate time.Time) error
	ReadMultiPartUpload(repoId, uploadId string) (*model.MultipartUpload, error)
	DeleteMultiPartUpload(repoId, uploadId string) error
}

func (index *DBIndex) writeEntryToWorkspace(tx store.RepoOperations, repo *model.Repo, branch, path string, entry *model.WorkspaceEntry) error {
	err := tx.WriteToWorkspacePath(branch, entry.ParentPath, path, entry)
	if err != nil {
		return err
	}
	if index.shouldPartiallyCommit(repo) {
		err = index.partialCommit(tx, branch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (index *DBIndex) shouldPartiallyCommit(repo *model.Repo) bool {
	chosen := rand.Float32()
	return chosen < DefaultPartialCommitRatio
}

func (index *DBIndex) partialCommit(tx store.RepoOperations, branch string) error {
	// see if we have any changes that weren't applied
	err := tx.LockWorkspace()
	if err != nil {
		return err
	}
	wsEntries, err := tx.ListWorkspace(branch)
	if err != nil {
		return err
	}
	if len(wsEntries) == 0 {
		return nil
	}

	// get branch info (including current workspace root)
	branchData, err := tx.ReadBranch(branch)
	if errors.Is(err, db.ErrNotFound) {
		return nil
	} else if err != nil {
		return err // unexpected error
	}

	// update the immutable Merkle tree, getting back a new tree
	tree := merkle.New(branchData.WorkspaceRoot, merkle.WithLogger(index.log()))
	tree, err = tree.Update(tx, wsEntries)
	if err != nil {
		return err
	}

	// clear workspace entries
	err = tx.ClearWorkspace(branch)
	if err != nil {
		return err
	}

	// update branch pointer to point at new workspace
	err = tx.WriteBranch(branch, &model.Branch{
		RepositoryId:  branchData.RepositoryId,
		Id:            branchData.Id,
		CommitId:      branchData.CommitId,
		CommitRoot:    branchData.CommitRoot,
		WorkspaceRoot: tree.Root(), // does this happen properly?
	})
	if err != nil {
		return err
	}

	// done!
	return nil
}

func gc(tx store.RepoOperations, addr string) {
	// TODO: impl? here?
}

type DBIndex struct {
	store       store.Store
	tsGenerator TimeGenerator

	partialCommitRatio float64
	ctx                context.Context
}

type Option func(index *DBIndex)

type TimeGenerator func() time.Time

// Option to initiate with
// when using this option timestamps will generate using the given time generator
// used for mocking and testing timestamps
func WithTimeGenerator(generator TimeGenerator) Option {
	return func(dbi *DBIndex) {
		dbi.tsGenerator = generator
	}
}

func WithContext(ctx context.Context) Option {
	return func(dbi *DBIndex) {
		dbi.ctx = ctx
	}
}

func WithPartialCommitRatio(ratio float64) Option {
	return func(dbi *DBIndex) {
		dbi.partialCommitRatio = ratio
	}
}

func NewDBIndex(db db.Database, opts ...Option) *DBIndex {
	kvi := &DBIndex{
		store:              store.NewDBStore(db),
		tsGenerator:        time.Now,
		partialCommitRatio: DefaultPartialCommitRatio,
		ctx:                context.Background(),
	}
	for _, opt := range opts {
		opt(kvi)
	}
	kvi.log().WithFields(logging.Fields{
		"partial_commit_ratio": kvi.partialCommitRatio,
	}).Info("initialized Metadata index")
	return kvi
}

func indexLogger(ctx context.Context) logging.Logger {
	return logging.FromContext(ctx).WithField("service_name", "index")
}

func (index *DBIndex) log() logging.Logger {
	return indexLogger(index.ctx)
}

// Business logic
func (index *DBIndex) WithContext(ctx context.Context) Index {
	return &DBIndex{
		store:              store.WithLogger(index.store, indexLogger(ctx)),
		tsGenerator:        index.tsGenerator,
		ctx:                ctx,
		partialCommitRatio: index.partialCommitRatio,
	}
}

type reference struct {
	commit   *model.Commit
	branch   *model.Branch
	isBranch bool
}

func (r *reference) String() string {
	if r.isBranch {
		return fmt.Sprintf("[branch='%s' -> commit='%s' -> root='%s']",
			r.branch.Id,
			r.commit.Address,
			r.commit.Tree)
	}
	return fmt.Sprintf("[commit='%s' -> root='%s']",
		r.commit.Address,
		r.commit.Tree)
}

func resolveRef(tx store.RepoOperations, ref string) (*reference, error) {
	// if this is not
	if ident.IsHash(ref) {
		// this looks like a straight up commit, let's see if it exists
		commit, err := tx.ReadCommit(ref)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			// got an error, we can't continue
			return nil, err
		} else if err == nil {
			// great, it's a commit, return it
			return &reference{
				commit: commit,
			}, nil
		}
	}
	// treat this as a branch name
	branch, err := tx.ReadBranch(ref)
	if err != nil {
		return nil, err
	}
	commit, err := tx.ReadCommit(branch.CommitId)
	if err != nil {
		return nil, err
	}

	return &reference{
		commit:   commit,
		branch:   branch,
		isBranch: true,
	}, nil
}

func (index *DBIndex) ReadObject(repoId, ref, path string, readUncommitted bool) (*model.Object, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
		ValidatePath(path),
	)
	if err != nil {
		return nil, err
	}

	obj, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		_, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}

		reference, err := resolveRef(tx, ref)
		if err != nil {
			return nil, err
		}
		var obj *model.Object

		if reference.isBranch && readUncommitted {
			we, err := tx.ReadFromWorkspace(reference.branch.Id, path)
			if errors.Is(err, db.ErrNotFound) {
				// not in workspace, let's try reading it from branch tree
				m := merkle.New(reference.branch.WorkspaceRoot, merkle.WithLogger(index.log()))
				obj, err = m.GetObject(tx, path)
				if err != nil {
					return nil, err
				}
				return obj, nil
			} else if err != nil {
				// an actual error has occurred, return it.
				index.log().WithError(err).Error("could not read from workspace")
				return nil, err
			}
			if we.Tombstone {
				// object was deleted deleted
				return nil, db.ErrNotFound
			}
			return tx.ReadObject(*we.EntryAddress)
		}
		// otherwise, read from commit
		m := merkle.New(reference.commit.Tree, merkle.WithLogger(index.log()))
		obj, err = m.GetObject(tx, path)
		if err != nil {
			return nil, err
		}
		return obj, nil
	}, db.ReadOnly())
	if err != nil {
		return nil, err
	}
	return obj.(*model.Object), nil
}

func (index *DBIndex) readEntry(tx store.RepoOperations, ref, path, typ string, readUncommitted bool) (*model.Entry, error) {
	var entry *model.Entry

	_, err := tx.ReadRepo()
	if err != nil {
		return nil, err
	}

	reference, err := resolveRef(tx, ref)
	if err != nil {
		return nil, err
	}
	root := reference.commit.Tree
	if reference.isBranch && readUncommitted {
		// try reading from workspace
		we, err := tx.ReadFromWorkspace(reference.branch.Id, path)

		// continue with we only if we got no error
		if err != nil {
			if !errors.Is(err, db.ErrNotFound) {
				return nil, err
			}
		} else {
			if we.Tombstone {
				// object was deleted deleted
				return nil, db.ErrNotFound
			}
			return &model.Entry{
				RepositoryId: we.RepositoryId,
				Name:         *we.EntryName,
				Address:      *we.EntryAddress,
				EntryType:    *we.EntryType,
				CreationDate: *we.EntryCreationDate,
				Size:         *we.EntrySize,
				Checksum:     *we.EntryChecksum,
			}, nil
		}
		root = reference.branch.WorkspaceRoot
	}

	m := merkle.New(root, merkle.WithLogger(index.log()))
	entry, err = m.GetEntry(tx, path, typ)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (index *DBIndex) ReadEntry(repoId, branch, path, typ string, readUncommitted bool) (*model.Entry, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path),
	)
	if err != nil {
		return nil, err
	}
	entry, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		return index.readEntry(tx, branch, path, typ, readUncommitted)
	}, db.ReadOnly())
	if err != nil {
		return nil, err
	}
	return entry.(*model.Entry), nil
}

func (index *DBIndex) ReadRootObject(repoId, ref string, readUncommitted bool) (*model.Root, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
	)
	if err != nil {
		return nil, err
	}
	root, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (i interface{}, err error) {
		_, err = tx.ReadRepo()
		if err != nil {
			return nil, err
		}
		reference, err := resolveRef(tx, ref)
		if err != nil {
			return nil, err
		}
		if reference.isBranch && readUncommitted {
			return tx.ReadRoot(reference.branch.WorkspaceRoot)
		}
		return tx.ReadRoot(reference.commit.Tree)
	})
	if err != nil {
		return nil, err
	}
	return root.(*model.Root), nil
}

func (index *DBIndex) ReadEntryTree(repoId, ref, path string, readUncommitted bool) (*model.Entry, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
		ValidatePath(path),
	)
	if err != nil {
		return nil, err
	}
	return index.ReadEntry(repoId, ref, path, model.EntryTypeTree, readUncommitted)
}

func (index *DBIndex) ReadEntryObject(repoId, ref, path string, readUncommitted bool) (*model.Entry, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
		ValidatePath(path),
	)
	if err != nil {
		return nil, err
	}
	return index.ReadEntry(repoId, ref, path, model.EntryTypeObject, readUncommitted)
}

func (index *DBIndex) WriteFile(repoId, branch, path string, entry *model.Entry, obj *model.Object) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path),
	)
	if err != nil {
		return err
	}
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		repo, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}

		err = tx.WriteObject(ident.Hash(obj), obj)
		if err != nil {
			index.log().WithError(err).Error("could not write object")
			return nil, err
		}
		err = index.writeEntryToWorkspace(tx, repo, branch, path, &model.WorkspaceEntry{
			RepositoryId:      repoId,
			BranchId:          branch,
			ParentPath:        pth.New(path, entry.EntryType).ParentPath(),
			Path:              path,
			EntryName:         &entry.Name,
			EntryAddress:      &entry.Address,
			EntryType:         &entry.EntryType,
			EntryCreationDate: &entry.CreationDate,
			EntrySize:         &entry.Size,
			EntryChecksum:     &entry.Checksum,
			Tombstone:         false,
		})
		if err != nil {
			index.log().WithError(err).Error("could not write workspace entry")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) WriteEntry(repoId, branch, path string, entry *model.Entry) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path))
	if err != nil {
		return err
	}
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		repo, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}
		err = index.writeEntryToWorkspace(tx, repo, branch, path, &model.WorkspaceEntry{
			RepositoryId:      repoId,
			BranchId:          branch,
			ParentPath:        pth.New(path, entry.EntryType).ParentPath(),
			Path:              path,
			EntryName:         &entry.Name,
			EntryAddress:      &entry.Address,
			EntryType:         &entry.EntryType,
			EntryCreationDate: &entry.CreationDate,
			EntrySize:         &entry.Size,
			EntryChecksum:     &entry.Checksum,
			Tombstone:         false,
		})
		if err != nil {
			index.log().WithError(err).Error("could not write workspace entry")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) WriteObject(repoId, branch, path string, object *model.Object) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path))
	if err != nil {
		return err
	}
	timestamp := index.tsGenerator()
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		addr := ident.Hash(object)
		err := tx.WriteObject(addr, object)
		if err != nil {
			return nil, err
		}
		repo, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}
		typ := model.EntryTypeObject
		p := pth.New(path, typ)
		entryName := pth.New(path, typ).BaseName()
		err = index.writeEntryToWorkspace(tx, repo, branch, path, &model.WorkspaceEntry{
			RepositoryId:      repoId,
			Path:              p.String(),
			ParentPath:        p.ParentPath(),
			BranchId:          branch,
			EntryName:         &entryName,
			EntryAddress:      &addr,
			EntryType:         &typ,
			EntryCreationDate: &timestamp,
			EntrySize:         &object.Size,
			EntryChecksum:     &object.Checksum,
			Tombstone:         false,
		})
		if err != nil {
			index.log().WithError(err).Error("could not write workspace entry")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) DeleteObject(repoId, branch, path string) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path))
	if err != nil {
		return err
	}
	ts := index.tsGenerator()
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		repo, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}
		/**
		handling 5 possible cases:
		* 1 object does not exist  - return error
		* 2 object exists only in workspace - remove from workspace
		* 3 object exists only in merkle - add tombstone
		* 4 object exists in workspace and in merkle - 2 + 3
		* 5 objects exists in merkle tombstone exists in workspace - return error
		*/
		notFoundCount := 0
		wsEntry, err := tx.ReadFromWorkspace(branch, path)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				notFoundCount += 1
			} else {
				return nil, err
			}
		}

		br, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		root := br.WorkspaceRoot
		m := merkle.New(root, merkle.WithLogger(index.log()))
		merkleEntry, err := m.GetEntry(tx, path, model.EntryTypeObject)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				notFoundCount += 1
			} else {
				return nil, err
			}
		}

		if notFoundCount == 2 {
			return nil, db.ErrNotFound
		}

		if wsEntry != nil {
			if wsEntry.Tombstone {
				return nil, db.ErrNotFound
			}
			err = tx.DeleteWorkspacePath(branch, path)
			if err != nil {
				return nil, err
			}
		}

		if merkleEntry != nil {
			typ := model.EntryTypeObject
			bname := pth.New(path, typ).BaseName()
			err = index.writeEntryToWorkspace(tx, repo, branch, path, &model.WorkspaceEntry{
				Path:              path,
				EntryName:         &bname,
				EntryCreationDate: &ts,
				EntryType:         &typ,
				Tombstone:         true,
			})
			if err != nil {
				index.log().WithError(err).Error("could not write workspace tombstone")
			}
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (index *DBIndex) ListBranchesByPrefix(repoId string, prefix string, amount int, after string) ([]*model.Branch, bool, error) {
	err := ValidateAll(
		ValidateRepoId(repoId))
	if err != nil {
		return nil, false, err
	}
	type result struct {
		hasMore bool
		results []*model.Branch
	}

	entries, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		// we're reading the repo to add it to this transaction's conflict range
		// but also to ensure it exists
		_, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}
		branches, hasMore, err := tx.ListBranches(prefix, amount, after)
		return &result{
			results: branches,
			hasMore: hasMore,
		}, err
	})
	if err != nil {
		index.log().WithError(err).Error("could not list branches")
		return nil, false, err
	}
	return entries.(*result).results, entries.(*result).hasMore, nil
}

func (index *DBIndex) ListObjectsByPrefix(repoId, ref, path, from string, results int, descend, readUncommitted bool) ([]*model.Entry, bool, error) {
	log := index.log().WithFields(logging.Fields{
		"from":    from,
		"descend": descend,
		"results": results,
	})
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
		ValidatePath(path),
	)
	if err != nil {
		return nil, false, err
	}
	type result struct {
		hasMore bool
		results []*model.Entry
	}
	entries, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		_, err := tx.ReadRepo()
		if err != nil {
			return nil, err
		}

		reference, err := resolveRef(tx, ref)
		if err != nil {
			return nil, err
		}

		root := reference.commit.Tree
		if reference.isBranch && readUncommitted {
			err := index.partialCommit(tx, reference.branch.Id) // block on this since we traverse the tree immediately after
			if err != nil {
				return nil, err
			}
			reference.branch, err = tx.ReadBranch(reference.branch.Id)
			if err != nil {
				return nil, err
			}
			root = reference.branch.WorkspaceRoot
		}

		tree := merkle.New(root, merkle.WithLogger(index.log()))
		res, hasMore, err := tree.PrefixScan(tx, path, from, results, descend)
		if err != nil {
			if !errors.Is(err, db.ErrNotFound) {
				log.WithError(err).Error("could not scan tree")
			}
			return nil, err
		}
		return &result{hasMore, res}, nil
	})
	if err != nil {
		return nil, false, err
	}
	return entries.(*result).results, entries.(*result).hasMore, nil
}

func (index *DBIndex) ResetBranch(repoId, branch string) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch))
	if err != nil {
		return err
	}
	// clear workspace, set branch workspace root back to commit root
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		err := tx.ClearWorkspace(branch)
		if err != nil {
			return nil, err
		}
		branchData, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		gc(tx, branchData.WorkspaceRoot)
		branchData.WorkspaceRoot = branchData.CommitRoot
		return nil, tx.WriteBranch(branch, branchData)
	})
	if err != nil {
		index.log().WithError(err).Error("could not reset branch")
	}
	return err
}

func (index *DBIndex) CreateBranch(repoId, branch, ref string) (*model.Branch, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(ref),
		ValidateRef(branch))
	if err != nil {
		return nil, err
	}
	branchData, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		// ensure it doesn't exist yet
		_, err := tx.ReadBranch(branch)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			index.log().WithError(err).Error("could not read branch")
			return nil, err
		}
		if err == nil {
			return nil, indexerrors.ErrBranchAlreadyExists
		}
		// read resolve reference
		reference, err := resolveRef(tx, ref)
		if err != nil {
			return nil, fmt.Errorf("could not read ref: %w", err)
		}
		branchData := &model.Branch{
			Id:            branch,
			CommitId:      reference.commit.Address,
			CommitRoot:    reference.commit.Tree,
			WorkspaceRoot: reference.commit.Tree,
		}
		return branchData, tx.WriteBranch(branch, branchData)
	})
	if err != nil {
		index.log().WithError(err).WithField("ref", ref).Error("could not create branch")
		return nil, err
	}
	return branchData.(*model.Branch), nil
}

func (index *DBIndex) GetBranch(repoId, branch string) (*model.Branch, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch))
	if err != nil {
		return nil, err
	}
	brn, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (i interface{}, err error) {
		return tx.ReadBranch(branch)
	}, db.ReadOnly())
	if err != nil {
		return nil, err
	}
	return brn.(*model.Branch), nil
}

func doCommitUpdates(tx store.RepoOperations, branchData *model.Branch, committer, message string, parents []string, metadata map[string]string, ts time.Time, index *DBIndex) (*model.Commit, error) {
	commit := &model.Commit{
		Tree:         branchData.WorkspaceRoot,
		Parents:      parents,
		Committer:    committer,
		Message:      message,
		CreationDate: ts,
		Metadata:     metadata,
	}
	commitAddr := ident.Hash(commit)
	commit.Address = commitAddr
	err := tx.WriteCommit(commitAddr, commit)
	if err != nil {
		index.log().WithError(err).Error("could not write commit")
		return nil, err
	}
	branchData.CommitId = commitAddr
	branchData.CommitRoot = commit.Tree
	branchData.WorkspaceRoot = commit.Tree

	return commit, tx.WriteBranch(branchData.Id, branchData)
}

func (index *DBIndex) Commit(repoId, branch, message, committer string, metadata map[string]string) (*model.Commit, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidateCommitMessage(message))
	if err != nil {
		return nil, err
	}
	ts := index.tsGenerator()
	commit, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		err := index.partialCommit(tx, branch)
		if err != nil {
			return nil, err
		}
		branchData, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		return doCommitUpdates(tx, branchData, committer, message, []string{branchData.CommitId}, metadata, ts, index)
	})
	if err != nil {
		return nil, err
	}
	return commit.(*model.Commit), nil
}

func (index *DBIndex) GetCommit(repoId, commitId string) (*model.Commit, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateCommitID(commitId))
	if err != nil {
		return nil, err
	}
	commit, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		return tx.ReadCommit(commitId)
	}, db.ReadOnly())
	if err != nil {
		return nil, err
	}
	return commit.(*model.Commit), nil
}

func (index *DBIndex) GetCommitLog(repoId, fromCommitId string, results int, after string) ([]*model.Commit, bool, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateCommitID(fromCommitId),
		ValidateOrEmpty(ValidateCommitID, after))

	type result struct {
		hasMore bool
		results []*model.Commit
	}
	if err != nil {
		return nil, false, err
	}
	res, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (i interface{}, err error) {
		commits, hasMore, err := dag.CommitScan(tx, fromCommitId, results, after)
		return &result{hasMore, commits}, err
	}, db.ReadOnly())
	if err != nil {
		index.log().WithError(err).WithField("from", fromCommitId).Error("could not read commits")
		return nil, false, err
	}
	return res.(*result).results, res.(*result).hasMore, nil
}

func (index *DBIndex) DeleteBranch(repoId, branch string) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch))
	if err != nil {
		return err
	}
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		branchData, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		err = tx.ClearWorkspace(branch)
		if err != nil {
			index.log().WithError(err).Error("could not clear workspace")
			return nil, err
		}
		gc(tx, branchData.WorkspaceRoot) // changes are destroyed here
		err = tx.DeleteBranch(branch)
		if err != nil {
			index.log().WithError(err).Error("could not delete branch")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) DiffWorkspace(repoId, branch string) (merkle.Differences, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch))
	if err != nil {
		return nil, err
	}
	res, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (i interface{}, err error) {
		err = index.partialCommit(tx, branch) // ensure all changes are reflected in the tree
		if err != nil {
			return nil, err
		}
		branch, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}

		diff, err := merkle.Diff(tx,
			merkle.New(branch.WorkspaceRoot, merkle.WithLogger(index.log())),
			merkle.New(branch.CommitRoot, merkle.WithLogger(index.log())),
			merkle.New(branch.CommitRoot, merkle.WithLogger(index.log())),
		)
		if err != nil {
			index.log().WithError(err).WithField("branch", branch).Error("diff workspace failed")
		}
		return diff, err
	})
	if err != nil {
		return nil, err
	}
	return res.(merkle.Differences), nil
}

func doDiff(tx store.RepoOperations, repoId, leftRef, rightRef string, isMerge bool, index *DBIndex) (merkle.Differences, error) {
	lRef, err := resolveRef(tx, leftRef)
	if err != nil {
		index.log().WithError(err).WithField("ref", leftRef).Error("could not resolve left ref")
		return nil, indexerrors.ErrBranchNotFound
	}

	rRef, err := resolveRef(tx, rightRef)
	if err != nil {
		index.log().WithError(err).WithField("ref", rRef).Error("could not resolve right ref")
		return nil, indexerrors.ErrBranchNotFound
	}

	commonCommits, err := dag.FindLowestCommonAncestor(tx, lRef.commit.Address, rRef.commit.Address)
	if err != nil {
		index.log().WithField("left", lRef).WithField("right", rRef).WithError(err).Error("could not find common commit")
		return nil, indexerrors.ErrNoMergeBase
	}
	if commonCommits == nil {
		index.log().WithField("left", lRef).WithField("right", rRef).Error("no common merge base found")
		return nil, indexerrors.ErrNoMergeBase
	}

	leftTree := lRef.commit.Tree
	if lRef.isBranch && !isMerge {
		leftTree = lRef.branch.WorkspaceRoot
	}
	rightTree := rRef.commit.Tree

	diff, err := merkle.Diff(tx,
		merkle.New(leftTree, merkle.WithLogger(index.log())),
		merkle.New(rightTree, merkle.WithLogger(index.log())),
		merkle.New(commonCommits.Tree, merkle.WithLogger(index.log())),
	)
	if err != nil {
		index.log().WithField("left", lRef).WithField("right", rRef).WithError(err).Error("could not calculate diff")
	}
	return diff, err
}

func (index *DBIndex) Diff(repoId, leftRef, rightRef string) (merkle.Differences, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(leftRef),
		ValidateRef(rightRef),
	)
	if err != nil {
		return nil, err
	}
	res, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (i interface{}, err error) {
		return doDiff(tx, repoId, leftRef, rightRef, false, index)
	})

	return res.(merkle.Differences), nil
}

func (index *DBIndex) RevertCommit(repoId, branch, commit string) error {
	log := index.log().WithFields(logging.Fields{
		"branch": branch,
		"commit": commit,
	})
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidateCommitID(commit))
	if err != nil {
		return err
	}
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		err := tx.ClearWorkspace(branch)
		if err != nil {
			log.WithError(err).Error("could not revert commit")
			return nil, err
		}
		commitData, err := tx.ReadCommit(commit)
		if err != nil {
			return nil, err
		}
		branchData, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		gc(tx, branchData.WorkspaceRoot)
		branchData.CommitId = commit
		branchData.CommitRoot = commitData.Tree
		branchData.WorkspaceRoot = commitData.Tree
		err = tx.WriteBranch(branch, branchData)
		if err != nil {
			log.WithError(err).Error("could not write branch")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) revertPath(repoId, branch, path, typ string) error {
	log := index.log().WithFields(logging.Fields{
		"branch": branch,
		"path":   path,
	})
	_, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		p := pth.New(path, typ)
		if p.IsRoot() {
			return nil, index.ResetBranch(repoId, branch)
		}

		err := index.partialCommit(tx, branch)
		if err != nil {
			log.WithError(err).Error("could not partially commit")
			return nil, err
		}
		branchData, err := tx.ReadBranch(branch)
		if err != nil {
			return nil, err
		}
		workspaceMerkle := merkle.New(branchData.WorkspaceRoot, merkle.WithLogger(index.log()))
		commitMerkle := merkle.New(branchData.CommitRoot, merkle.WithLogger(index.log()))
		var workspaceEntry *model.WorkspaceEntry
		commitEntry, err := commitMerkle.GetEntry(tx, path, typ)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				// remove all changes under path
				pathEntry, err := workspaceMerkle.GetEntry(tx, path, typ)
				if err != nil {
					return nil, err
				}
				workspaceEntry = &model.WorkspaceEntry{
					RepositoryId:      repoId,
					BranchId:          branch,
					ParentPath:        p.ParentPath(),
					Path:              path,
					EntryName:         &pathEntry.Name,
					EntryAddress:      &pathEntry.Address,
					EntryType:         &pathEntry.EntryType,
					EntryCreationDate: &pathEntry.CreationDate,
					EntrySize:         &pathEntry.Size,
					EntryChecksum:     &pathEntry.Checksum,
					Tombstone:         true,
				}
			} else {
				log.WithError(err).Error("could not get entry")
				return nil, err
			}
		} else {
			workspaceEntry = &model.WorkspaceEntry{
				RepositoryId:      repoId,
				BranchId:          branch,
				ParentPath:        p.ParentPath(),
				Path:              path,
				EntryName:         &commitEntry.Name,
				EntryAddress:      &commitEntry.Address,
				EntryType:         &commitEntry.EntryType,
				EntryCreationDate: &commitEntry.CreationDate,
				EntrySize:         &commitEntry.Size,
				EntryChecksum:     &commitEntry.Checksum,
			}
		}
		commitEntries := []*model.WorkspaceEntry{workspaceEntry}
		workspaceMerkle, err = workspaceMerkle.Update(tx, commitEntries)
		if err != nil {
			log.WithError(err).Error("could not update Merkle tree")
			return nil, err
		}

		// update branch workspace pointer to point at new workspace
		err = tx.WriteBranch(branch, &model.Branch{
			Id:            branch,
			CommitId:      branchData.CommitId,
			CommitRoot:    branchData.CommitRoot,
			WorkspaceRoot: workspaceMerkle.Root(),
		})

		if err != nil {
			log.WithError(err).Error("could not write branch")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) RevertPath(repoId, branch, path string) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path))
	if err != nil {
		return err
	}
	return index.revertPath(repoId, branch, path, model.EntryTypeTree)
}

func (index *DBIndex) RevertObject(repoId, branch, path string) error {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(branch),
		ValidatePath(path))
	if err != nil {
		return err
	}
	return index.revertPath(repoId, branch, path, model.EntryTypeObject)
}

func (index *DBIndex) Merge(repoId, source, destination, userId string) (merkle.Differences, error) {
	err := ValidateAll(
		ValidateRepoId(repoId),
		ValidateRef(source),
		ValidateRef(destination),
	)
	if err != nil {
		return nil, err
	}
	ts := index.tsGenerator()
	var mergeOperations merkle.Differences
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		// check that destination has no uncommitted changes
		destinationBranch, err := tx.ReadBranch(destination)
		if err != nil {
			index.log().WithError(err).WithField("destination", destination).Warn(" branch " + destination + " not found")
			return nil, indexerrors.ErrBranchNotFound
		}
		l, err := tx.ListWorkspace(destination)
		if err != nil {
			index.log().WithError(err).WithField("destination", destination).Warn(" branch " + destination + " workspace not found")
			return nil, err
		}
		if destinationBranch.CommitRoot != destinationBranch.WorkspaceRoot || len(l) > 0 {
			return nil, indexerrors.ErrDestinationNotCommitted
		}
		// compute difference
		df, err := doDiff(tx, repoId, source, destination, true, index)
		if err != nil {
			return nil, err
		}
		var isConflict bool
		for _, dif := range df {
			if dif.Direction == merkle.DifferenceDirectionConflict {
				isConflict = true
			}
			if dif.Direction != merkle.DifferenceDirectionRight {
				mergeOperations = append(mergeOperations, dif)
			}
		}
		if isConflict {
			return nil, indexerrors.ErrMergeConflict
		}
		// update destination with source changes
		var wsEntries []*model.WorkspaceEntry
		sourceBranch, err := tx.ReadBranch(source)
		if err != nil {
			index.log().WithError(err).Fatal("failed reading source branch") // failure to read a branch that was read before fatal
			return nil, err
		}
		for _, dif := range mergeOperations {
			var e *model.Entry
			m := merkle.New(sourceBranch.WorkspaceRoot, merkle.WithLogger(index.log()))
			if dif.Type != merkle.DifferenceTypeRemoved {
				e, err = m.GetEntry(tx, dif.Path, dif.PathType)
				if err != nil {
					index.log().WithError(err).Fatal("failed reading entry")
					return nil, err
				}
			} else {
				e = new(model.Entry)
				p := strings.Split(dif.Path, "/")
				e.Name = p[len(p)-1]
				e.EntryType = dif.PathType
			}
			w := new(model.WorkspaceEntry)
			w.EntryType = &e.EntryType
			w.EntryAddress = &e.Address
			w.EntryName = &e.Name
			w.EntryChecksum = &e.Checksum
			w.EntryCreationDate = &e.CreationDate
			w.EntrySize = &e.Size
			w.Path = dif.Path
			w.Tombstone = (dif.Type == merkle.DifferenceTypeRemoved)
			wsEntries = append(wsEntries, w)
		}

		destinationRoot := merkle.New(destinationBranch.CommitRoot, merkle.WithLogger(index.log()))
		newRoot, err := destinationRoot.Update(tx, wsEntries)
		if err != nil {
			index.log().WithError(err).Fatal("failed updating merge destination")
			return nil, indexerrors.ErrMergeUpdateFailed
		}
		destinationBranch.CommitRoot = newRoot.Root()
		destinationBranch.WorkspaceRoot = newRoot.Root()

		// read commits for each branch in our merge.
		// check which parents is older by searching the other parents using our DAG
		// 1. read commits
		// 2. use iterator with the first commit to lookup the other commit
		branches := []*model.Branch{sourceBranch, destinationBranch}
		commits := make([]*model.Commit, len(branches))
		for i, branch := range branches {
			var err error
			commits[i], err = tx.ReadCommit(branch.CommitId)
			if err != nil {
				index.log().WithError(err).Error("failed read commit")
				return nil, fmt.Errorf("missing commit: %w", err)
			}
		}
		parent1Commit := commits[0]
		parent2Commit := commits[1]
		iter := dag.NewCommitIterator(tx, parent2Commit.Address)
		parent2Older := true
		for iter.Next() {
			if iter.Value().Address == parent1Commit.Address {
				parent2Older = false
				break
			}
		}
		if iter.Err() != nil {
			index.log().WithError(err).Error("failed while lookup parent relation")
			return nil, fmt.Errorf("failed to scan parent commits: %w", err)
		}
		if !parent2Older {
			commits[0], commits[1] = commits[1], commits[0]
		}
		parents := []string{commits[0].Address, commits[1].Address}

		commitMessage := "Merge branch " + source + " into " + destination
		_, err = doCommitUpdates(tx, destinationBranch, userId, commitMessage, parents, make(map[string]string), ts, index)
		if err != nil {
			index.log().WithError(err).WithFields(logging.Fields{
				"source":      source,
				"destination": destination,
				"userId":      userId,
				"parents":     parents,
			}).Error("commit merge branch")
			return nil, fmt.Errorf("failed to commit updates: %w", err)
		}
		return mergeOperations, nil
	})
	// ErrMergeConflict is the only error that will report the merge operations made so far
	if err != nil && err != indexerrors.ErrMergeConflict {
		return nil, err
	}
	return mergeOperations, err
}

func (index *DBIndex) CreateRepo(repoId, bucketName, defaultBranch string) error {
	err := ValidateAll(
		ValidateRepoId(repoId))
	if err != nil {
		return err
	}

	creationDate := index.tsGenerator()

	repo := &model.Repo{
		Id:               repoId,
		StorageNamespace: bucketName,
		CreationDate:     creationDate,
		DefaultBranch:    defaultBranch,
	}

	// create repository, an empty commit and tree, and the default branch
	_, err = index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		// make sure this repo doesn't already exist
		_, err := tx.ReadRepo()
		if err == nil {
			// couldn't verify this bucket doesn't yet exist
			return nil, indexerrors.ErrRepoExists
		} else if !errors.Is(err, db.ErrNotFound) {
			index.log().WithError(err).Error("could not read repo")
			return nil, err // error reading the repo
		}

		err = tx.WriteRepo(repo)
		if err != nil {
			return nil, err
		}

		// write empty tree
		err = tx.WriteRoot(ident.Empty(), &model.Root{
			RepositoryId: repoId,
			Address:      ident.Empty(),
			CreationDate: creationDate,
			Size:         0,
		})
		if err != nil {
			return nil, err
		}

		commit := &model.Commit{
			RepositoryId: repoId,
			Tree:         ident.Empty(),
			Committer:    "",
			Message:      "",
			CreationDate: creationDate,
			Parents:      []string{},
			Metadata:     make(map[string]string),
		}
		commitId := ident.Hash(commit)
		commit.Address = commitId
		err = tx.WriteCommit(commitId, commit)
		if err != nil {
			index.log().WithError(err).Error("could not write initial commit")
			return nil, err
		}
		err = tx.WriteBranch(repo.DefaultBranch, &model.Branch{
			Id:            repo.DefaultBranch,
			CommitId:      commitId,
			CommitRoot:    commit.Tree,
			WorkspaceRoot: commit.Tree,
		})
		if err != nil {
			index.log().WithError(err).Error("could not write branch")
		}
		return nil, err
	})
	return err
}

func (index *DBIndex) ListRepos(amount int, after string) ([]*model.Repo, bool, error) {
	type result struct {
		repos   []*model.Repo
		hasMore bool
	}
	res, err := index.store.Transact(func(tx store.ClientOperations) (interface{}, error) {
		repos, hasMore, err := tx.ListRepos(amount, after)
		return &result{
			repos:   repos,
			hasMore: hasMore,
		}, err
	}, db.ReadOnly())
	if err != nil {
		index.log().WithError(err).Error("could not list repos")
		return nil, false, err
	}
	return res.(*result).repos, res.(*result).hasMore, nil
}

func (index *DBIndex) GetRepo(repoId string) (*model.Repo, error) {
	err := ValidateAll(
		ValidateRepoId(repoId))
	if err != nil {
		return nil, err
	}
	repo, err := index.store.Transact(func(tx store.ClientOperations) (interface{}, error) {
		return tx.ReadRepo(repoId)
	}, db.ReadOnly())
	if err != nil {
		return nil, err
	}
	return repo.(*model.Repo), nil
}

func (index *DBIndex) DeleteRepo(repoId string) error {
	err := ValidateAll(
		ValidateRepoId(repoId))
	if err != nil {
		return err
	}
	_, err = index.store.Transact(func(tx store.ClientOperations) (interface{}, error) {
		_, err := tx.ReadRepo(repoId)
		if err != nil {
			return nil, err
		}
		err = tx.DeleteRepo(repoId)
		if err != nil {
			index.log().WithError(err).Error("could not delete repo")
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (index *DBIndex) CreateDedupEntryIfNone(repoId string, dedupId string, objName string) (string, error) {
	objectId, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		dedupObj, err := tx.GetObjectDedup(dedupId)
		if err == nil {
			return dedupObj.PhysicalAddress, nil
		} else if errors.Is(err, db.ErrNotFound) {
			d := &model.ObjectDedup{RepositoryId: repoId, PhysicalAddress: objName, DedupId: dedupId}
			err = tx.WriteObjectDedup(d)
			if err != nil {
				index.log().WithError(err).Error("failed writing dedup record")
			}
			return objName, err
		} else {
			index.log().WithError(err).Error("Error reading  dedup record")
			return objName, err
		}
	})
	val, ok := objectId.(string)
	if ok {
		return val, err
	} else {
		return objName, err
	}
}

func (index *DBIndex) CreateMultiPartUpload(repoId, uploadId, path, objectName string, creationDate time.Time) error {
	_, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		m := &model.MultipartUpload{RepositoryId: repoId, UploadId: uploadId, Path: path, CreationDate: creationDate, PhysicalAddress: objectName}
		err := tx.WriteMultipartUpload(m)
		return nil, err
	})
	return err
}

func (index *DBIndex) ReadMultiPartUpload(repoId, uploadId string) (*model.MultipartUpload, error) {
	multi, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		m, err := tx.ReadMultipartUpload(uploadId)
		if err != nil {
			index.log().WithError(err).Error("failed reading MultiPart record")
		}
		return m, err
	})

	if err == nil {
		return multi.(*model.MultipartUpload), err
	} else {
		return nil, err
	}
}

func (index *DBIndex) DeleteMultiPartUpload(repoId, uploadId string) error {
	_, err := index.store.RepoTransact(repoId, func(tx store.RepoOperations) (interface{}, error) {
		err := tx.DeleteMultipartUpload(uploadId)
		return nil, err
	})
	return err
}
