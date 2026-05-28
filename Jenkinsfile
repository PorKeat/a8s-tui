pipeline {
    // Run directly on the Jenkins server host machine (no Docker required)
    agent any

    environment {
        // These credentials must be configured in Jenkins
        GITHUB_TOKEN = credentials('github-token-cli')
        NODE_AUTH_TOKEN = credentials('npm-token')
    }

    stages {


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
