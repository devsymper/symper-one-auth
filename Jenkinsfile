/* groovylint-disable LineLength */
pipeline {

    agent any

    environment {
        SERVICE_NAME = "symper-one-auth.symper.vn"
        HOST_NAME = 'symper-one-auth-forward.symper.vn'

        APP_NAME = sh(script: "echo $SERVICE_NAME | cut -d'.' -f1", returnStdout: true).trim()
        Author_Name=sh(script: "git show -s --pretty=%ae", returnStdout: true).trim()
        BRANCH_NAME = "${GIT_BRANCH.split('/').size() > 1 ? GIT_BRANCH.split('/')[1..-1].join('/') : GIT_BRANCH}"
        COMMIT_MESSAGE = sh(script: "git log --format=%B -n 1", returnStdout: true).trim()

        PROJECT_NAME = 'symper'
        MSTEAMS_WEBHOOK=credentials('ms_teams_webhook')
        MSTEAMS_DEPLOY_WEBHOOK = credentials('ms_teams_deploy_notify')

        // Build image config options
        DOCKER_BUILDKIT = 1
        BUILDKIT_INLINE_CACHE = 1
        REGISTRY_HOST = 'localhost:5000'
    }

    stages {
        stage ("quality control") {
            when {
                allOf {
                    expression { env.BRANCH_NAME == 'dev' }
                    triggeredBy 'UserIdCause'
                }
            }
            environment {
                SERVICE_ENV = 'test'

                SSH_HOST = "10.20.166.15"
                POSTGRES_HOST = "10.20.166.52"
                POSTGRES_DB = "symper_one_auth"

                CLICKHOUSE_DB = "symper_one_auth"
                CLICKHOUSE_HOST = "10.20.166.52"

                BUILD_VERSION = 'latest'
                IMAGE_TAG_NAME = sh(script: "echo ${PROJECT_NAME}/${SERVICE_NAME}:${BUILD_VERSION}", returnStdout: true).trim()
            }
            stages {
                stage('build') {
                    steps {
                        withCredentials([
                            usernamePassword(
                                credentialsId: 'harbor_registry',
                                passwordVariable: 'DOCKER_REGISTRY_PWD',
                                usernameVariable: 'DOCKER_REGISTRY_USER'
                            )
                        ]) {
                            script {
                                echo "IMAGE_TAG_NAME: $IMAGE_TAG_NAME"

                                try {
                                    sh '''
                                        set -e
                                        echo $DOCKER_REGISTRY_PWD | docker login -u $DOCKER_REGISTRY_USER --password-stdin localhost:5000
                                    '''
                                } catch (Exception e) {
                                    echo "Docker login failed: ${e.getMessage()}"
                                    error 'Stopping pipeline due to Docker login failure'
                                }

                                sh 'docker build -t localhost:5000/$IMAGE_TAG_NAME .'
                                sh "docker push localhost:5000/${IMAGE_TAG_NAME}"
                                sh "docker image rm localhost:5000/${IMAGE_TAG_NAME}"
                            }
                        }
                    }
                }
                stage('checkin deployment state') {
					steps {
						withCredentials([
                            usernamePassword(credentialsId: 'ssh_qc_vps', passwordVariable: 'USER_PASS', usernameVariable: 'USER_NAME')
                        ]) {
							script {
								sshagent(['ssh_qc_key']) {
									try {
										env.CURRENT_ROLE = sh(returnStdout: true,
                                                            script: "ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST 'echo -e \'$USER_PASS\' | sudo -S kubectl get services --field-selector metadata.name=\"$APP_NAME\" -o jsonpath={.items[0].spec.selector.role}'").trim()
                                    } catch (Exception e) {
										env.CURRENT_ROLE = ''
                                        echo "$e"
                                    }
                                }
                                sh "echo role ${env.CURRENT_ROLE}"
                                if ("$env.CURRENT_ROLE" == '' || "$env.CURRENT_ROLE" == 'green') {
									env.CURRENT_ROLE = 'green'
                                    env.TARGET_ROLE = 'blue'

                                } else {
									env.TARGET_ROLE = 'green'
                                }
                            }
                        }
                    }
                }
                stage("deploy to k8s") {
                    steps{
                        withCredentials([
                            usernamePassword(credentialsId: 'dev_database', passwordVariable: 'POSTGRES_PASS', usernameVariable: 'POSTGRES_USER'),
                            usernamePassword(credentialsId: 'ssh_qc_vps', passwordVariable: 'USER_PASS', usernameVariable: 'USER_NAME')
                        ]) {
							sh 'chmod +x deploy/scripts/*'
                            sshagent(['ssh_qc_key']) {
								script {
									echo "WITH TARGET_ROLE: $TARGET_ROLE"
                                    echo "START DEPLOY TO K8S"
                                    try {
										sh '''
                                            set -e
                                            ./deploy/scripts/update_manifests.sh
                                        '''
                                    } catch (Exception e) {
										echo "Error in update_manifests.sh: ${e.getMessage()}"
                                        error "Stopping pipeline due to failure in update_manifests.sh"
                                    }

                                    try {
										sh '''
                                            set -e
                                            ./deploy/scripts/deploy_service.sh
                                        '''
                                    } catch (Exception e) {
										echo "Error in deploy_service.sh: ${e.getMessage()}"
                                        error "Stopping pipeline due to failure in deploy_service.sh"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        stage('production') {
            when {
                allOf {
                    expression { env.GIT_BRANCH?.contains('tags') }
                    triggeredBy cause: 'UserIdCause'
                }
            }
            environment {
				SERVICE_ENV = 'prod'
                SSH_HOST = '10.20.166.246'

                POSTGRES_HOST = "10.20.166.193"
                POSTGRES_DB = "symper_one_auth"

                CLICKHOUSE_DB = "default"
                CLICKHOUSE_HOST = "10.20.166.166"
                CACHE_HOST = "10.20.166.90"

                BUILD_VERSION = sh(returnStdout:  true, script: 'git tag --sort=-creatordate | head -n 1').trim()
                IMAGE_TAG_NAME = sh(script: "echo ${PROJECT_NAME}/${SERVICE_NAME}:${BUILD_VERSION}", returnStdout: true).trim()
            }
            stages {
                stage('build') {
                    steps {
                        withCredentials([
                            usernamePassword(
                                credentialsId: 'harbor_registry',
                                passwordVariable: 'DOCKER_REGISTRY_PWD',
                                usernameVariable: 'DOCKER_REGISTRY_USER'
                            )
                        ]) {
                            script {
                                echo "IMAGE_TAG_NAME: $IMAGE_TAG_NAME"

                                try {
                                    sh '''
                                        set -e
                                        echo $DOCKER_REGISTRY_PWD | docker login -u $DOCKER_REGISTRY_USER --password-stdin localhost:5000
                                    '''
                                } catch (Exception e) {
                                    echo "Docker login failed: ${e.getMessage()}"
                                    error 'Stopping pipeline due to Docker login failure'
                                }

                                sh 'docker build -t localhost:5000/$IMAGE_TAG_NAME .'
                                sh "docker push localhost:5000/${IMAGE_TAG_NAME}"
                                sh "docker image rm localhost:5000/${IMAGE_TAG_NAME}"
                            }
                        }
                    }
                }
                stage('checkin deployment state') {
					steps {
						withCredentials([
                            usernamePassword(credentialsId: 'ssh_prod_vps', passwordVariable: 'USER_PASS', usernameVariable: 'USER_NAME')
                        ]) {
							script {
								sshagent(['prod_ssh_key']) {
									try {
										env.CURRENT_ROLE = sh(returnStdout: true,
                                                            script: "ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST 'echo -e \'$USER_PASS\' | sudo -S kubectl get services --field-selector metadata.name=\"$APP_NAME\" -o jsonpath={.items[0].spec.selector.role}'").trim()
                                    } catch (Exception e) {
										env.CURRENT_ROLE = ''
                                        echo "$e"
                                    }
                                }
                                sh "echo role ${env.CURRENT_ROLE}"
                                if ("$env.CURRENT_ROLE" == '' || "$env.CURRENT_ROLE" == 'green') {
									env.CURRENT_ROLE = 'green'
                                    env.TARGET_ROLE = 'blue'
                                } else {
									env.TARGET_ROLE = 'green'
                                }
                            }
                        }
                    }
                }
                stage('deploy to k8s') {
					steps {
						withCredentials([
                            usernamePassword(credentialsId: 'symper_one_auth_db', passwordVariable: 'POSTGRES_PASS', usernameVariable: 'POSTGRES_USER'),
                            usernamePassword(credentialsId: 'clickhouse_data_io', passwordVariable: 'CLICKHOUSE_PASS', usernameVariable: 'CLICKHOUSE_USER'),
                            usernamePassword(credentialsId: 'ssh_prod_vps', passwordVariable: 'USER_PASS', usernameVariable: 'USER_NAME')
                        ]) {
							sh 'chmod +x deploy/scripts/*'
                            sshagent(['prod_ssh_key']) {
								script {
									echo "WITH TARGET_ROLE: $TARGET_ROLE"
                                    echo "START DEPLOY TO K8S"
                                    try {
										sh '''
                                            set -e
                                            ./deploy/scripts/update_manifests.sh
                                        '''
                                    } catch (Exception e) {
										echo "Error in update_manifests.sh: ${e.getMessage()}"
                                        error "Stopping pipeline due to failure in update_manifests.sh"
                                    }

                                    try {
										sh '''
                                            set -e
                                            ./deploy/scripts/deploy_service.sh
                                        '''
                                    } catch (Exception e) {
										echo "Error in deploy_service.sh: ${e.getMessage()}"
                                        error "Stopping pipeline due to failure in deploy_service.sh"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    // Send notify deploy
    post {
        success {
            script {
                // Only push notification when deploy successfully
                if (env.TARGET_ROLE?.trim()) {
                    office365ConnectorSend(
                        webhookUrl: env.MSTEAMS_DEPLOY_WEBHOOK,
                        message: """
                            \n✅ *Build Success*
                            \n- Job: ${env.JOB_NAME}
                            \n- Build Number: #${env.BUILD_NUMBER}
                            \n- Branch: ${env.GIT_BRANCH}
                            \n- Build URL: ${env.BUILD_URL}
                        """,
                        status: 'Success',
                        adaptiveCards: true
                    )
                }
            }
        }
        failure {
            script {
                // Only send notifications when a deployment is triggered manually or via pipeline
                // - Check that the branch is either a development branch or a tagged release
                // - Ensure TARGET_ROLE is set (Jenkins sets this only during actual deployments, not during auto-scans)
                if (env.GIT_BRANCH?.contains('tags') || env.GIT_BRANCH?.contains('dev')) {
                    office365ConnectorSend(
                        webhookUrl: env.MSTEAMS_DEPLOY_WEBHOOK,
                        message: """
                            \n❌ *Build Failed*
                            \n- Job: ${env.JOB_NAME}
                            \n- Build Number: #${env.BUILD_NUMBER}
                            \n- Branch: ${env.GIT_BRANCH}
                            \n- Build URL: ${env.BUILD_URL}
                        """,
                        status: 'Failure',
                        adaptiveCards: true
                    )
                }
            }
        }
    }
}
