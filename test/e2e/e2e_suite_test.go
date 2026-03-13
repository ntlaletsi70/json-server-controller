/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ntlaletsi70/json-server-controller/test/utils"
)

var (
	// Optional Environment Variables:
	// - CERT_MANAGER_INSTALL_SKIP=true
	skipCertManagerInstall = os.Getenv("CERT_MANAGER_INSTALL_SKIP") == "true"

	// detect existing cert manager
	isCertManagerAlreadyInstalled = false

	// image used for e2e testing
	projectImage = "example.com/json-server-controller:v0.0.1"
)

// ----------------------------------------------------
// Test Suite
// ----------------------------------------------------

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting json-server-controller integration test suite\n")
	RunSpecs(t, "e2e suite")
}

// ----------------------------------------------------
// Before Suite
// ----------------------------------------------------

var _ = BeforeSuite(func() {

	//------------------------------------------------
	// Build image using Mage (instead of Make)
	//------------------------------------------------

	By("building the manager(Operator) image")

	cmd := exec.Command("mage", "build:image")
	cmd.Env = append(os.Environ(), "IMG="+projectImage)

	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(
		HaveOccurred(),
		"Failed to build the manager(Operator) image",
	)

	//------------------------------------------------
	// Load image into Kind
	//------------------------------------------------

	By("loading the manager(Operator) image on Kind")

	err = utils.LoadImageToKindClusterWithName(projectImage)
	ExpectWithOffset(1, err).NotTo(
		HaveOccurred(),
		"Failed to load the manager(Operator) image into Kind",
	)

	//------------------------------------------------
	// Cert Manager Setup
	//------------------------------------------------

	if !skipCertManagerInstall {

		By("checking if cert manager is installed already")

		isCertManagerAlreadyInstalled = utils.IsCertManagerCRDsInstalled()

		if !isCertManagerAlreadyInstalled {

			_, _ = fmt.Fprintf(GinkgoWriter, "Installing CertManager...\n")

			Expect(utils.InstallCertManager()).To(
				Succeed(),
				"Failed to install CertManager",
			)

		} else {

			_, _ = fmt.Fprintf(
				GinkgoWriter,
				"WARNING: CertManager is already installed. Skipping installation...\n",
			)

		}
	}
})

// ----------------------------------------------------
// After Suite
// ----------------------------------------------------

var _ = AfterSuite(func() {

	if !skipCertManagerInstall && !isCertManagerAlreadyInstalled {

		_, _ = fmt.Fprintf(GinkgoWriter, "Uninstalling CertManager...\n")

		utils.UninstallCertManager()
	}
})
