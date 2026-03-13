//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Img           = getEnv("IMG", "json-server-controller:2h")
	ContainerTool = getEnv("CONTAINER_TOOL", "docker")
	LocalBin      = "./bin"

	controllerGenVersion = "v0.17.2"
	kustomizeVersion     = "v5.5.0"
)

type Build mg.Namespace
type Dev mg.Namespace

// ----------------------------------------------------
// TOOL BOOTSTRAP
// ----------------------------------------------------

func EnsureTools() error {

	step("Ensuring required tools exist")

	if err := os.MkdirAll(LocalBin, 0755); err != nil {
		return err
	}

	if err := ensureControllerGen(); err != nil {
		return err
	}

	if err := ensureKustomize(); err != nil {
		return err
	}

	info("Tools ready")

	return nil
}

func ensureControllerGen() error {

	path := filepath.Join(LocalBin, "controller-gen")

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	step("Installing controller-gen")

	abs, err := absBin()
	if err != nil {
		return err
	}

	return sh.RunV(
		"env",
		"GOBIN="+abs,
		"go",
		"install",
		"sigs.k8s.io/controller-tools/cmd/controller-gen@"+controllerGenVersion,
	)
}
func ensureKustomize() error {

	path := filepath.Join(LocalBin, "kustomize")

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	step("Installing kustomize")

	abs, err := absBin()
	if err != nil {
		return err
	}

	return sh.RunV(
		"env",
		"GOBIN="+abs,
		"go",
		"install",
		"sigs.k8s.io/kustomize/kustomize/v5@"+kustomizeVersion,
	)
}

//
// DEVELOPMENT
//

func Manifests() error {

	mg.Deps(EnsureTools)

	step("Generating CRDs and RBAC")

	return sh.RunV(
		LocalBin+"/controller-gen",
		"rbac:roleName=manager-role",
		"crd",
		"webhook",
		"paths=./...",
		"output:crd:artifacts:config=config/crd/bases",
	)
}

func Generate() error {

	mg.Deps(EnsureTools)

	step("Generating deepcopy code")

	return sh.RunV(
		LocalBin+"/controller-gen",
		"object:headerFile=hack/boilerplate.go.txt",
		"paths=./...",
	)
}

func Fmt() error {

	step("Formatting Go code")

	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		fail("go fmt failed")
		return err
	}

	info("Formatting complete")

	return nil
}

func Vet() error {

	step("Running go vet")

	return sh.RunV("go", "vet", "./...")
}

//
// BUILD
//

func (Build) Manager() error {

	mg.Deps(Manifests, Generate, Fmt, Vet)

	step("Building manager binary")

	if err := sh.RunV(
		"go",
		"build",
		"-o",
		"bin/manager",
		"cmd/main.go",
	); err != nil {
		return err
	}

	info("Binary built → bin/manager")

	return nil
}

func (Build) Image() error {

	mg.Deps(Build{}.Manager)

	step("Building container image")

	return sh.RunV(
		ContainerTool,
		"build",
		"-t",
		Img,
		".",
	)
}

func (Build) Push() error {

	mg.Deps(Build{}.Image)

	step("Pushing container image")

	return sh.RunV(
		ContainerTool,
		"push",
		Img,
	)
}

//
// KUBERNETES
//

func Install() error {

	mg.Deps(Manifests)

	step("Installing CRDs")

	return sh.RunV(
		"bash",
		"-c",
		LocalBin+"/kustomize build config/crd | kubectl apply -f -",
	)
}

func Deploy() error {

	mg.Deps(Manifests)

	step("Updating controller image")

	if err := sh.RunV(
		"bash",
		"-c",
		"cd config/manager && ../../bin/kustomize edit set image controller="+Img,
	); err != nil {
		return err
	}

	step("Deploying controller")

	if err := sh.RunV(
		"bash",
		"-c",
		LocalBin+"/kustomize build config/default | kubectl apply -f -",
	); err != nil {
		return err
	}

	info("Controller deployed successfully")

	return nil
}

//
// DEVELOPMENT RUN
//

func (Dev) Run() error {

	step("Starting controller locally")

	return sh.RunV(
		"go",
		"run",
		"./cmd/main.go",
		"--leader-elect=false",
	)
}

//
// HELPERS
//

//
// HELPERS
//

func absBin() (string, error) {
	return filepath.Abs(LocalBin)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func step(msg string) {
	fmt.Printf("\n\033[36m▶ %s\033[0m\n", msg)
}

func info(msg string) {
	fmt.Printf("\033[32m✔ %s\033[0m\n", msg)
}

func warn(msg string) {
	fmt.Printf("\033[33m⚠ %s\033[0m\n", msg)
}

func fail(msg string) {
	fmt.Printf("\033[31m✖ %s\033[0m\n", msg)
}
