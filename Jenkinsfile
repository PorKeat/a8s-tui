pipeline {
    // Run directly on the Jenkins server host machine (no Docker required)
    agent any

    parameters {
        string(name: 'RELEASE_VERSION', defaultValue: '', description: 'Leave blank for a normal test build. To publish a release, enter the version (e.g., 1.0.0)')
    }

    environment {
        // These credentials must be configured in Jenkins
        GITHUB_TOKEN = credentials('github-token-cli')
        NODE_AUTH_TOKEN = credentials('npm-token')
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
                    
                    # Install GoReleaser if missing
                    if ! command -v goreleaser >/dev/null 2>&1; then
                        echo "Downloading and installing GoReleaser..."
                        wget -qO- https://github.com/goreleaser/goreleaser/releases/download/v1.26.2/goreleaser_Linux_x86_64.tar.gz | tar -xz -C $HOME/.local/bin/ goreleaser
                    fi
                '''
            }
        }

        stage('Build & Test') {
            // Always run build and test on every push to ensure code is healthy
            steps {
                sh 'go build .'
                sh 'go test ./...'
            }
        }

        stage('Publish Release (Manual)') {
            // This stage only runs if you type a version number into Jenkins when you click Build
            when { 
                expression { 
                    return params.RELEASE_VERSION != null && params.RELEASE_VERSION.trim() != ''
                } 
            }
            stages {
                stage('Tag Repository') {
                    steps {
                        sh '''
                            VERSION=${RELEASE_VERSION}
                            if [[ $VERSION != v* ]]; then
                                VERSION="v$VERSION"
                            fi
                            
                            # Configure Git so it can tag
                            git config user.email "jenkins@a8s.com"
                            git config user.name "Jenkins Auto-Publisher"
                            
                            # Create the tag locally
                            git tag $VERSION
                            
                            # Push the tag to GitHub using the injected GITHUB_TOKEN
                            git push https://${GITHUB_TOKEN}@github.com/ITProfessional-Gen01/a8s-cli.git $VERSION
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
                        sh '''
                            VERSION=${RELEASE_VERSION}
                            if [[ $VERSION == v* ]]; then
                                VERSION=${VERSION#v}
                            fi
                            
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
