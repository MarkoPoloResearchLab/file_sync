package sync

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

type SyncResult struct {
	ChangedFileCount int
	ActionCounters   map[string]int
	Diff3Available   bool
}

// RunSync performs a bidirectional synchronization between two roots.
func RunSync(options Options, logger *zap.Logger) (SyncResult, error) {
	var result SyncResult
	result.ActionCounters = map[string]int{
		"A<-B (create)": 0,
		"B<-A (create)": 0,
		"merge(seed)":   0,
		"merge(3way)":   0,
		"merge(2way)":   0,
		"equal":         0,
		"absent":        0,
	}

	store, state, err := createOrOpenStateStore(options.StateDirectory)
	if err != nil {
		if logger != nil {
			logger.Error("open state store", zap.Error(err))
		}
		return result, err
	}

	relativeSet := map[string]struct{}{}

	err = filepath.WalkDir(options.RootAPath, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			rel, _ := filepath.Rel(options.RootAPath, currentPath)
			if rel != "." && shouldIgnorePath(rel, options.IgnorePathPrefixes) {
				return filepath.SkipDir
			}
			return nil
		}
		fileName := d.Name()
		if shouldIgnoreName(fileName, options.IgnoreFileNames) {
			return nil
		}
		rel, _ := filepath.Rel(options.RootAPath, currentPath)
		matchRel, _ := filepath.Match(options.IncludeGlob, filepath.ToSlash(rel))
		matchName, _ := filepath.Match(options.IncludeGlob, fileName)
		if matchRel || matchName {
			relativeSet[filepath.ToSlash(rel)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		if logger != nil {
			logger.Error("walk root A", zap.String("root", options.RootAPath), zap.Error(err))
		}
		return result, err
	}

	err = filepath.WalkDir(options.RootBPath, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			rel, _ := filepath.Rel(options.RootBPath, currentPath)
			if rel != "." && shouldIgnorePath(rel, options.IgnorePathPrefixes) {
				return filepath.SkipDir
			}
			return nil
		}
		fileName := d.Name()
		if shouldIgnoreName(fileName, options.IgnoreFileNames) {
			return nil
		}
		rel, _ := filepath.Rel(options.RootBPath, currentPath)
		matchRel, _ := filepath.Match(options.IncludeGlob, filepath.ToSlash(rel))
		matchName, _ := filepath.Match(options.IncludeGlob, fileName)
		if matchRel || matchName {
			relativeSet[filepath.ToSlash(rel)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		if logger != nil {
			logger.Error("walk root B", zap.String("root", options.RootBPath), zap.Error(err))
		}
		return result, err
	}

	relativeList := make([]string, 0, len(relativeSet))
	for rel := range relativeSet {
		relativeList = append(relativeList, rel)
	}
	sort.Strings(relativeList)

	diff3Path, lookupErr := exec.LookPath("diff3")
	diff3Available := lookupErr == nil
	result.Diff3Available = diff3Available

	for _, relativePath := range relativeList {
		changed, tag, procErr := processSingleFile(relativePath, options, store, state, diff3Path, logger)
		if procErr != nil {
			if logger != nil {
				logger.Error("process file", zap.String("path", relativePath), zap.Error(procErr))
			}
			return result, procErr
		}
		if changed {
			result.ChangedFileCount++
		}
		result.ActionCounters[tag] = result.ActionCounters[tag] + 1
	}

	if err := store.save(state); err != nil {
		if logger != nil {
			logger.Error("save state", zap.Error(err))
		}
		return result, err
	}
	return result, nil
}

func readAll(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func writeAllEnsure(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func processSingleFile(relativePath string, options Options, store *stateStore, state *syncState, diff3Path string, logger *zap.Logger) (bool, string, error) {
	pathA := filepath.Join(options.RootAPath, relativePath)
	pathB := filepath.Join(options.RootBPath, relativePath)

	_, statAErr := os.Stat(pathA)
	_, statBErr := os.Stat(pathB)

	existsA := !errors.Is(statAErr, fs.ErrNotExist)
	existsB := !errors.Is(statBErr, fs.ErrNotExist)

	entry := state.FileEntry[relativePath]

	if existsA && !existsB {
		content, readErr := readAll(pathA)
		if readErr != nil {
			if logger != nil {
				logger.Error("read file", zap.String("path", pathA), zap.Error(readErr))
			}
			return false, "", readErr
		}
		if err := writeAllEnsure(pathB, content); err != nil {
			if logger != nil {
				logger.Error("write file", zap.String("path", pathB), zap.Error(err))
			}
			return false, "", err
		}
		hexDigest, ancErr := store.ensureAncestorStored(content)
		if ancErr != nil {
			if logger != nil {
				logger.Error("store ancestor", zap.Error(ancErr))
			}
			return false, "", ancErr
		}
		state.FileEntry[relativePath] = stateEntry{AncestorHex: hexDigest}
		return true, "B<-A (create)", nil
	}

	if existsB && !existsA {
		content, readErr := readAll(pathB)
		if readErr != nil {
			if logger != nil {
				logger.Error("read file", zap.String("path", pathB), zap.Error(readErr))
			}
			return false, "", readErr
		}
		if err := writeAllEnsure(pathA, content); err != nil {
			if logger != nil {
				logger.Error("write file", zap.String("path", pathA), zap.Error(err))
			}
			return false, "", err
		}
		hexDigest, ancErr := store.ensureAncestorStored(content)
		if ancErr != nil {
			if logger != nil {
				logger.Error("store ancestor", zap.Error(ancErr))
			}
			return false, "", ancErr
		}
		state.FileEntry[relativePath] = stateEntry{AncestorHex: hexDigest}
		return true, "A<-B (create)", nil
	}

	if !existsA && !existsB {
		delete(state.FileEntry, relativePath)
		return false, "absent", nil
	}

	contentA, readAErr := readAll(pathA)
	if readAErr != nil {
		if logger != nil {
			logger.Error("read file", zap.String("path", pathA), zap.Error(readAErr))
		}
		return false, "", readAErr
	}
	contentB, readBErr := readAll(pathB)
	if readBErr != nil {
		if logger != nil {
			logger.Error("read file", zap.String("path", pathB), zap.Error(readBErr))
		}
		return false, "", readBErr
	}

	if bytesEqual(contentA, contentB) {
		if entry.AncestorHex == "" {
			hexDigest, ancErr := store.ensureAncestorStored(contentA)
			if ancErr != nil {
				if logger != nil {
					logger.Error("store ancestor", zap.Error(ancErr))
				}
				return false, "", ancErr
			}
			state.FileEntry[relativePath] = stateEntry{AncestorHex: hexDigest}
		}
		return false, "equal", nil
	}

	var baseBytes []byte
	if entry.AncestorHex != "" {
		loaded, ancErr := store.ancestorBytes(entry.AncestorHex)
		if ancErr == nil {
			baseBytes = loaded
		}
	}

	if baseBytes == nil {
		infoA, _ := os.Stat(pathA)
		infoB, _ := os.Stat(pathB)
		modA := modtimeSeconds(infoA)
		modB := modtimeSeconds(infoB)
		var merged []byte

		if options.CreateBackupsOnWrite {
			_ = copyFile(pathA, pathA+".bak.a")
			_ = copyFile(pathB, pathB+".bak.b")
		}

		if absFloat64(modA-modB) <= options.ConflictMtimeEpsilonSeconds {
			merged = mergeWithMarkers(contentA, contentB)
		} else if modA > modB {
			merged = contentA
		} else {
			merged = contentB
		}

		if err := writeAllEnsure(pathA, merged); err != nil {
			if logger != nil {
				logger.Error("write file", zap.String("path", pathA), zap.Error(err))
			}
			return false, "", err
		}
		if err := writeAllEnsure(pathB, merged); err != nil {
			if logger != nil {
				logger.Error("write file", zap.String("path", pathB), zap.Error(err))
			}
			return false, "", err
		}
		hexDigest, ancErr := store.ensureAncestorStored(merged)
		if ancErr != nil {
			if logger != nil {
				logger.Error("store ancestor", zap.Error(ancErr))
			}
			return false, "", ancErr
		}
		state.FileEntry[relativePath] = stateEntry{AncestorHex: hexDigest}
		return true, "merge(seed)", nil
	}

	if options.CreateBackupsOnWrite {
		_ = copyFile(pathA, pathA+".bak.a")
		_ = copyFile(pathB, pathB+".bak.b")
	}

	merged, diffUsed := mergeThreeWay(mergeInputs{
		BaseBytes:  baseBytes,
		SideABytes: contentA,
		SideBBytes: contentB,
		Diff3Path:  diff3Path,
	})

	if err := writeAllEnsure(pathA, merged); err != nil {
		if logger != nil {
			logger.Error("write file", zap.String("path", pathA), zap.Error(err))
		}
		return false, "", err
	}
	if err := writeAllEnsure(pathB, merged); err != nil {
		if logger != nil {
			logger.Error("write file", zap.String("path", pathB), zap.Error(err))
		}
		return false, "", err
	}

	hexDigest, ancErr := store.ensureAncestorStored(merged)
	if ancErr != nil {
		if logger != nil {
			logger.Error("store ancestor", zap.Error(ancErr))
		}
		return false, "", ancErr
	}
	state.FileEntry[relativePath] = stateEntry{AncestorHex: hexDigest}

	if diffUsed {
		return true, "merge(3way)", nil
	}
	return true, "merge(2way)", nil
}

func bytesEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func copyFile(fromPath string, toPath string) error {
	data, err := os.ReadFile(fromPath)
	if err != nil {
		return err
	}
	return writeAllEnsure(toPath, data)
}

func modtimeSeconds(info fs.FileInfo) float64 {
	if info == nil {
		return 0
	}
	return float64(info.ModTime().UnixNano()) / float64(time.Second)
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
