#!/usr/bin/env groovy

@Library(['product-pipelines-shared-library', 'conjur-enterprise-sharedlib']) _

def productName = 'Conjur Kubernetes CSI Provider'
def productTypeName = 'Conjur Enterprise'

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([
    'conjur-enterprise/conjur-authn-k8s-client',
    'conjur-enterprise/conjur-api-go'
  ])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { infrapool, sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    env.INFRAPOOL_PRODUCT_NAME = "${productName}"
    env.INFRAPOOL_DD_PRODUCT_TYPE_NAME = "${productTypeName}"

    // Scan the images before promoting
    def scans = [:]

    scans["Scan main Docker image"] = {
      runSecurityScans(infrapool,
        image: "registry.tld/conjur-k8s-csi-provider:${sourceVersion}",
        buildMode: params.MODE,
        branch: env.BRANCH_NAME,
        arch: 'linux/amd64')
    }

    scans["Scan RedHat Docker image"] = {
      runSecurityScans(infrapool,
            image: "registry.tld/conjur-k8s-csi-provider-redhat:${sourceVersion}",
            buildMode: params.MODE,
            branch: env.BRANCH_NAME,
            arch: 'linux/amd64')
    }

    parallel scans

    // Pull existing images from internal registry in order to promote
    infrapool.agentSh """
      export PATH="release-tools/bin:${PATH}"
      docker pull registry.tld/conjur-k8s-csi-provider:${sourceVersion}
      docker pull registry.tld/conjur-k8s-csi-provider-redhat:${sourceVersion}
      # Promote source version to target version.
      summon --environment release bin/publish --promote --source ${sourceVersion} --target ${targetVersion}
    """
  }

  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    parameterizedCron(getDailyCronString("%NIGHTLY=true"))
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()

    // Values to direct scan results to the right place in DefectDojo
    INFRAPOOL_PRODUCT_NAME = "${productName}"
    INFRAPOOL_DD_PRODUCT_TYPE_NAME = "${productTypeName}"
  }

  parameters {
    booleanParam(name: 'NIGHTLY', defaultValue: false, description: 'Run CSI Provider tests against all supported platforms: Kubernetes, Openshift Current/Oldest')
    booleanParam(name: 'TEST_OCP_NEXT', defaultValue: false, description: 'Run CSI Provider tests against next Openshift version')
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }

    stage('Scan for internal URLs') {
      steps {
        script {
          detectInternalUrls()
        }
      }
    }

    stage('Get InfraPool ExecutorV2 Agent(s)') {
      steps{
        script {
          // Request ExecutorV2 agents for 1 hour(s)
          INFRAPOOL_EXECUTORV2_AGENTS = getInfraPoolAgent(type: "ExecutorV2", quantity: 1, duration: 1)
          INFRAPOOL_EXECUTORV2_AGENT_0 = INFRAPOOL_EXECUTORV2_AGENTS[0]
          infrapool = infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0, {})
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        script {
          updateVersion(infrapool, "CHANGELOG.md", "${BUILD_NUMBER}")
        }
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
          // Copy the vendor directory onto infrapool
          infrapool.agentPut from: "vendor", to: "${WORKSPACE}"
          infrapool.agentPut from: "go.*", to: "${WORKSPACE}"
          infrapool.agentPut from: "/root/go", to: "/var/lib/jenkins/"
        }
      }
    }

    stage('Build Docker image') {
      steps {
        script {
          infrapool.agentSh 'bin/build'
        }
      }
    }

    // Required for scanning the images.
    stage('Push images to internal registry') {
      steps {
        script {
          infrapool.agentSh './bin/publish --internal'
        }
      }
    }

    stage('Scan Docker Image') {
      parallel {
        stage("Scan main Docker Image") {
          steps {
            script {
              VERSION = infrapool.agentSh(returnStdout: true, script: 'cat VERSION')
              runSecurityScans(infrapool,
                image: "registry.tld/conjur-k8s-csi-provider:${VERSION}",
                buildMode: params.MODE,
                branch: env.BRANCH_NAME,
                arch: 'linux/amd64')
            }
          }
        }
        stage("Scan RedHat Docker image") {
          steps {
            script {
              VERSION = infrapool.agentSh(returnStdout: true, script: 'cat VERSION')
              runSecurityScans(infrapool,
                image: "registry.tld/conjur-k8s-csi-provider-redhat:${VERSION}",
                buildMode: params.MODE,
                branch: env.BRANCH_NAME,
                arch: 'linux/amd64')
            }
          }
        }
      }
    }

    stage('Helm tests'){
      parallel {
        stage('Helm unittest') {
          steps { script { infrapool.agentSh 'bin/test_helm_unit' } }
        }
        stage('Helm lint') {
          steps { script { infrapool.agentSh 'bin/test_helm_schema' } }
        }
      }
    }

    stage('Validate log messages') {
      steps { validateLogMessages() }
    }

    stage('Unit tests'){
      steps { script { infrapool.agentSh 'bin/test_unit' } }
    }

    stage ("E2E Test") {
      steps {
        script {
          infrapool.agentSh 'bin/test_e2e'
        }
      }
    }

    stage("E2E Test (Openshift - Current)") {
      when {
        expression { params.NIGHTLY }
      }
      steps {
        script { infrapool.agentSh 'bin/test_e2e openshift current' }
      }
    }

    stage("E2E Test (Openshift - Oldest)") {
      when {
        expression { params.NIGHTLY }
      }
      steps {
        script { infrapool.agentSh 'bin/test_e2e openshift oldest' }
      }
    }

    stage("E2E Test (Openshift - Next)") {
      when {
        expression { params.TEST_OCP_NEXT }
      }
      steps {
        script { infrapool.agentSh 'bin/test_e2e openshift next' }
      }
    }

    stage('Package artifacts') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          infrapool.agentSh 'bin/package_helm'
        }
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          release(infrapool) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations

            // Copy any artifacts to assetDirectory to attach them to the Github release
            infrapool.agentSh "cp -r helm-artifacts/*.tgz ${assetDirectory}"

            // Create Go application SBOM using the go.mod version for the golang container image
            infrapool.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/conjur-k8s-csi-provider/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
            // Create Go module SBOM
            infrapool.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
            infrapool.agentSh 'bin/publish --edge'
          }
        }
      }
    }
  }
  post {
    always {
      releaseInfraPoolAgent()
    }
  }
}
