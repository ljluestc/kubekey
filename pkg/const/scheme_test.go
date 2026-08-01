/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package _const

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestSchemeRegistersBatchV1CronJob asserts the kubekey controller-runtime
// scheme recognizes batch/v1 CronJob objects. This is a regression guard
// for issue #2086: any servicemesh / jaeger manifests ported to v4 must
// declare apiVersion: batch/v1, the only API version that survived the
// Kubernetes 1.25 removal of batch/v1beta1.
func TestSchemeRegistersBatchV1CronJob(t *testing.T) {
	if !Scheme.IsVersionRegistered(batchv1.SchemeGroupVersion) {
		t.Fatalf("scheme should register batch/v1; got nothing")
	}
	gvk := schema.GroupVersionKind{
		Group:   batchv1.GroupName,
		Version: "v1",
		Kind:    "CronJob",
	}
	if !Scheme.Recognizes(gvk) {
		t.Fatalf("scheme should recognize batch/v1 CronJob; missing GVK: %s", gvk)
	}
	// Confirm CronJob resolves to a non-nil concrete type.
	if _, err := Scheme.New(gvk); err != nil {
		t.Fatalf("scheme should construct a batch/v1 CronJob: %v", err)
	}
}

// TestNoBatchV1Beta1InTree walks the embedded builtin/plugins/config
// manifests and fails if any file declares the removed apiVersion:
// batch/v1beta1. The test pairs with the make lint-regression-batch-v1
// target so the guard fires both from `go test ./...` and from an
// explicit CI step.
func TestNoBatchV1Beta1InTree(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test source location; guard cannot run")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))

	roots := []string{
		filepath.Join(repoRoot, "builtin"),
		filepath.Join(repoRoot, "plugins"),
		filepath.Join(repoRoot, "config"),
	}

	scanned := 0
	for _, root := range roots {
		// Roots are not strictly required today — none of builtin, plugins,
		// or config currently ships servicemesh manifests. The guard is
		// future-facing, so a missing root is reported as a no-op rather
		// than a failure.
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !manifestExt(path) {
				return nil
			}
			scanned++
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(body), "batch/v1beta1") {
				t.Errorf("forbidden apiVersion batch/v1beta1 found in %s", path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", root, err)
		}
	}
	t.Logf("scanned %d manifest files across %d roots for batch/v1beta1", scanned, len(roots))
}

func manifestExt(path string) bool {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".yaml"),
		strings.HasSuffix(lower, ".yml"),
		strings.HasSuffix(lower, ".tmpl"),
		strings.HasSuffix(lower, ".j2"):
		return true
	}
	return false
}
