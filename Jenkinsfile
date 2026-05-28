pipeline {
    // Run directly on the Jenkins server host machine (no Docker required)
    agent any

    parameters {
        choice(name: 'RELEASE_TYPE', choices: ['None', 'Patch (0.0.x)', 'Minor (0.x.0)', 'Major (x.0.0)'], description: 'Select the type of release. Select "None" to just run normal tests without publishing.')
    }

    environment {
        // These credentials must be configured in Jenkins
        GITHUB_TOKEN = credentials('github-token-cli')
        NODE_AUTH_TOKEN = credentials('npm-token')
        
        // Keycloak Secrets to bake into the binary
        KEYCLOAK_URL = credentials('keycloak-url')
        KEYCLOAK_REALM = credentials('keycloak-realm')
        KEYCLOAK_CLIENT_ID = credentials('keycloak-client-id')
        KEYCLOAK_CLIENT_SECRET = credentials('keycloak-client-secret')
        KEYCLOAK_REDIRECT_URL = credentials('keycloak-redirect-url')
        
        // Prepend our local installation directories to the PATH
        PATH = "${HOME}/.local/go/bin:${HOME}/.local/node/bin:${HOME}/.local/bin:${env.PATH}"
    }

    stages {
        stage('Install Dependencies') {
            steps {
                sh '''
                    mkdir -p $HOME/.local/bin
                    
                    # Install Go if missing
                    if ! command -v go >/dev/null 2>&1; then
                        echo "Downloading and installing Go..."
                        wget -qO- https://go.dev/dl/go1.22.3.linux-amd64.tar.gz | tar -xz -C $HOME/.local/
                    fi
                    
                    # Install Node.js if missing
                    if ! command -v npm >/dev/null 2>&1; then
                        echo "Downloading and installing Node.js..."
                        wget -qO- https://nodejs.org/dist/v20.14.0/node-v20.14.0-linux-x64.tar.gz | tar -xz -C $HOME/.local/
                        rm -rf $HOME/.local/node || true
                        mv $HOME/.local/node-v20.14.0-linux-x64 $HOME/.local/node
                    fi
                    
                    # Install GoReleaser if missing (v2 required for version: 2 config)
                    if ! command -v goreleaser >/dev/null 2>&1; then
                        echo "Downloading and installing GoReleaser v2..."
                        wget -qO- https://github.com/goreleaser/goreleaser/releases/download/v2.0.1/goreleaser_Linux_x86_64.tar.gz | tar -xz -C $HOME/.local/bin/ goreleaser
                    fi
                '''
            }
        }

        stage('Build & Test') {
            // Always run build and test on every push to ensure code is healthy
            steps {
                sh '''#!/bin/bash
                    # Inject Keycloak secrets directly into the Go source code before compiling
                    cat <<EOF > config/config_defaults.go
package config

var (
	DefaultKeycloakURL          = "${KEYCLOAK_URL}"
	DefaultKeycloakRealm        = "${KEYCLOAK_REALM}"
	DefaultKeycloakClientID     = "${KEYCLOAK_CLIENT_ID}"
	DefaultKeycloakClientSecret = "${KEYCLOAK_CLIENT_SECRET}"
	DefaultKeycloakRedirectURL  = "${KEYCLOAK_REDIRECT_URL}"
)
EOF
                '''
                sh 'go build .'
                sh 'go test ./...'
            }
        }

        stage('Publish Release') {
            // This stage only runs if you selected Patch, Minor, or Major
            when { 
                expression { 
                    return params.RELEASE_TYPE != 'None'
                } 
            }
            stages {
                stage('Tag Repository') {
                    steps {
                        sh '''#!/bin/bash
                            # Find the absolute highest tag in the entire repository
                            LATEST=$(git tag -l | sort -V | tail -n 1)
                            if [ -z "$LATEST" ]; then
                                LATEST="v0.0.0"
                            fi
                            VERSION=${LATEST#v}
                            
                            # Split version into Major, Minor, Patch
                            IFS='.' read -ra ADDR <<< "$VERSION"
                            MAJOR=${ADDR[0]:-0}
                            MINOR=${ADDR[1]:-0}
                            PATCH=${ADDR[2]:-0}
                            
                            # Auto-increment the correct number
                            if [ "$RELEASE_TYPE" = "Patch (0.0.x)" ]; then
                                PATCH=$((PATCH + 1))
                            elif [ "$RELEASE_TYPE" = "Minor (0.x.0)" ]; then
                                MINOR=$((MINOR + 1))
                                PATCH=0
                            elif [ "$RELEASE_TYPE" = "Major (x.0.0)" ]; then
                                MAJOR=$((MAJOR + 1))
                                MINOR=0
                                PATCH=0
                            fi
                            
                            NEW_VERSION="v$MAJOR.$MINOR.$PATCH"
                            echo "🚀 Auto-bumping version from $LATEST to $NEW_VERSION"
                            
                            # Save the new version to a file so the NPM stage knows what it is
                            echo "$NEW_VERSION" > .release_version
                            
                            # Configure Git so it can tag
                            git config user.email "jenkins@a8s.com"
                            git config user.name "Jenkins Auto-Publisher"
                            
                            # Create the tag locally
                            git tag $NEW_VERSION
                            
                            # Push the tag to GitHub using the injected GITHUB_TOKEN
                            git push https://${GITHUB_TOKEN}@github.com/ITProfessional-Gen01/a8s-cli.git $NEW_VERSION
                        '''
                    }
                }
                stage('Release Go Binaries') {
                    steps {
                        sh 'goreleaser release --clean'
                    }
                }
                
                stage('Publish to NPM') {
                    steps {
                        sh 'npm ci'
                        sh '''#!/bin/bash
                            # Read the version we just created
                            VERSION=$(cat .release_version)
                            VERSION=${VERSION#v}
                            
                            # Copy all binaries from dist into bin so they get packaged into NPM
                            mkdir -p bin
                            cp dist/a8s-cli_darwin_amd64_v1/a8s-cli bin/a8s-cli_darwin_amd64 || true
                            cp dist/a8s-cli_darwin_arm64/a8s-cli bin/a8s-cli_darwin_arm64 || true
                            cp dist/a8s-cli_linux_amd64_v1/a8s-cli bin/a8s-cli_linux_amd64 || true
                            cp dist/a8s-cli_linux_arm64/a8s-cli bin/a8s-cli_linux_arm64 || true
                            cp dist/a8s-cli_windows_amd64_v1/a8s-cli.exe bin/a8s-cli_windows_amd64.exe || true
                            
                            # Update package.json version
                            npm version $VERSION --no-git-tag-version
                            
                            # Publish to NPM registry
                            npm publish
                        '''
                    }
                }
            }
        }
    }
}
