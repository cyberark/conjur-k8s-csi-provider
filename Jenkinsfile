#!/usr/bin/env groovy

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

    // Pull existing images from internal registry in order to promote
    infrapool.agentSh """
      export PATH="release-tools/bin:${PATH}"
      docker pull registry.tld/conjur-k8s-csi-provider:${sourceVersion}
      # Promote source version to target version.
      summon bin/publish --promote --source ${sourceVersion} --target ${targetVersion}
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
    cron(getDailyCronString())
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
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
          withCredentials([usernamePassword(credentialsId: 'jenkins_ci_token', usernameVariable: 'GITHUB_USER', passwordVariable: 'TOKEN')]) {
            sh './bin/updateGoDependencies.sh -g "${WORKSPACE}/go.mod"'
          }
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

    stage('Scan Docker Image') {
      parallel {
        stage("Scan Docker Image for fixable issues") {
          steps {
            // Adding the false parameter to scanAndReport causes trivy to
            // ignore vulnerabilities for which no fix is available. We'll
            // only fail the build if we can actually fix the vulnerability
            // right now.
            scanAndReport(infrapool, 'conjur-k8s-csi-provider:latest', "HIGH", false)
          }
        }
        stage("Scan Docker image for total issues") {
          steps {
            // By default, trivy includes vulnerabilities with no fix. We
            // want to know about that ASAP, but they shouldn't cause a
            // build failure until we can do something about it. This call
            // to scanAndReport should always be left as "NONE"
            scanAndReport(infrapool, "conjur-k8s-csi-provider:latest", "NONE", true)
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

    stage('Unit tests'){
      steps { script { infrapool.agentSh 'bin/test_unit' } }
    }

    stage('E2E tests') {
      steps { script { infrapool.agentSh 'bin/test_e2e' } }
    }

    // Allows for the promotion of images.
    stage('Push images to internal registry') {
      steps {
        script {
          infrapool.agentSh './bin/publish --internal'
        }
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
