pipeline {
    // Run directly on the Jenkins server host machine (no Docker required)
    agent any

    environment {
        // These credentials must be configured in Jenkins
        GITHUB_TOKEN = credentials('github-token-cli')
        NODE_AUTH_TOKEN = credentials('npm-token')
    }

    stages {
        stage('Initialize Git') {
            steps {
                // Ensure all tags are fetched so GoReleaser and npm version work
                sh 'git fetch --tags'
            }
        }

        stage('Build & Test (Main Branch)') {
            when { branch 'main' }
            steps {
                sh 'go build .'
                sh 'go test ./...'
            }
        }

        stage('Publish Release (Tags Only)') {
            when { buildingTag() }
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
