version_settings(constraint='>=0.22.1')

# Extensions
load('ext://git_resource', 'git_checkout')
load('ext://namespace', 'namespace_create', 'namespace_inject')

namespace_create('enginecare')

docker_build(
    'enginecare-image',
    '.',
    dockerfile='deployments/api/Dockerfile',
    ssh=['default'],
)

k8s_yaml('deployments/api/api.yaml')

k8s_resource(
    'enginecare',
    port_forwards="4000:4000",
    labels=["application"],
)
