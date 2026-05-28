pipeline {
    // Run directly on the Jenkins server host machine (no Docker required)
    agent any

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

        stage('Publish Release (Tags Only)') {
            // This runs a git command to see if the current commit is a tag. 
            // It works flawlessly in standard pipelines.
            when { 
                expression { 
                    return sh(script: 'git describe --exact-match --tags HEAD > /dev/null 2>&1', returnStatus: true) == 0 
                } 
            }
            stages {
                stage('Release Go Binaries') {
                    steps {
                        sh 'goreleaser release --clean'
                    }
                }
                
                stage('Publish to NPM') {
                    steps {
                        sh 'npm ci'
                        sh '''
                            # Get the current tag and strip the 'v'
                            VERSION=$(git describe --tags --abbrev=0)
                            VERSION=${VERSION#v}
                            
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
