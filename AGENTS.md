# KubeKey — Agent Guide
 
 ## Project overview
 
 KubeKey is an open-source lightweight task flow execution tool. It provides a flexible way to install and manage Kubernetes clusters (see `README.md`).
 
 The primary CLI binary is `kk`.
 
 ## Tech stack
 
 - **Language**: Go (module `github.com/kubesphere/kubekey/v4`).
 - **CLI**: `cobra` (root command is constructed in `cmd/kk/app`).
 - **Kubernetes/controller tooling**: controller-runtime, controller-gen, kustomize (see `Makefile`).
 - **Linting**: golangci-lint (configured in `.github/workflows/golangci-lint.yaml` and `.golangci.yaml`).
 
 ## Repo layout
 
 - `cmd/kk/kubekey.go`: `kk` CLI entrypoint.
 - `cmd/kk/app/`: CLI command tree.
 - `api/`: Kubernetes APIs and separate module (`api/go.mod`); root `go.mod` uses `replace` to point at `./api`.
 - `pkg/`: core implementation.
 - `config/`: Helm chart / manifests for deploying KubeKey into a Kubernetes cluster.
 - `builtin/`: built-in task templates and defaults.
 - `docs/`: documentation.
 - `hack/`: scripts (including `hack/downloadKubekey.sh`).
 
 ## Build and test commands
 
 The repository uses a large `Makefile` as the main entry point.
 
 - **Build `kk`**:
 
 ```bash
 make kk
 ```
 
 The binary is written to `_output/bin/kk` (see `Makefile`).
 
 - **Run tests**:
 
 ```bash
 make test
 ```
 
 - **Lint**:
 
 ```bash
 make lint
 ```
 
 - **Update module files**:
 
 ```bash
 make generate-modules
 ```
 
 ## Usage (high level)
 
 `README.md` contains end-to-end examples. Common entry points:
 
 - Install to a Kubernetes cluster via Helm:
 
 ```bash
 helm upgrade --install --create-namespace -n kubekey-system kubekey config/kubekey
 ```
 
 - Create inventory/config and create a cluster:
 
 ```bash
 kk create inventory
 kk create config --with-kubernetes v1.33.1
 kk create cluster -i inventory.yaml -c config.yaml
 ```
 
 ## Development conventions
 
 - Prefer `make` targets; CI uses `make generate-modules`, `make verify`, and `make test`.
 - Go versions are referenced in multiple places (root `go.mod`, `Makefile`, GitHub workflows).
 
 ## Security considerations
 
 - KubeKey is used to provision and manage infrastructure. Treat inventory/config files as sensitive.
 - Avoid committing credentials (SSH passwords/keys, kubeconfig, registry credentials, etc.).

## Task Implementation
1. **Analyze Requirements**: Refer to `README.md` for detailed feature specifications and system design.
2. **Implementation**: Modify source code in the respective directories (e.g., `src/`, `internal/`).
3. **Verification**: Run provided build and test commands (see above) to ensure correctness.
4. **Push Changes**:
   - Commit changes: `git commit -m "feat: implement <feature>"`
   - Push to remote: `git push origin <branch-name>`
