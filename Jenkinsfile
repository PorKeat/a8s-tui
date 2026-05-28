pipeline {
    agent none

    environment {
        // These credentials must be configured in Jenkins
        GITHUB_TOKEN = credentials('github-token-cli')
        NODE_AUTH_TOKEN = credentials('npm-token')
    }

    stages {
        stage('Initialize Git') {
            agent any
            steps {
                // Ensure all tags are fetched so GoReleaser and npm version work
                sh 'git fetch --tags'
            }
        }

        stage('Build & Test (Main Branch)') {
            when { branch 'main' }
            agent {
                docker { 
                    image 'golang:latest' 
                }
            }
            steps {
                sh 'go build .'
                sh 'go test ./...'
            }
        }

        stage('Publish Release (Tags Only)') {
            when { buildingTag() }
            stages {
                stage('Release Go Binaries') {
                    agent {
                        docker { 
                            image 'goreleaser/goreleaser:v2' 
                        }
                    }
                    steps {
                        sh 'goreleaser release --clean'
                    }
                }
                
                stage('Publish to NPM') {
                    agent {
                        docker { 
                            image 'node:20' 
                        }
                    }
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
