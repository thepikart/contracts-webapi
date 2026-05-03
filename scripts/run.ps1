param (
    $command
)

if (-not $command)  {
    $command = "start"
}

$ProjectRoot = "${PSScriptRoot}/.."

$env:CONTRACTS_API_ENVIRONMENT="Development"
$env:CONTRACTS_API_PORT="8080"
$env:CONTRACTS_API_MONGODB_USERNAME="root"
$env:CONTRACTS_API_MONGODB_PASSWORD="neUhaDnes"

function mongo {
    docker compose --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/contracts-api-service
        }
        finally {
            mongo down
        }
    }
    "test" {
        go test -v ./...
    }
    "mongo" {
        mongo up
    }
    "docker" {
        docker build -t thepikart/contracts-webapi:local-build -f ${ProjectRoot}/build/docker/Dockerfile .
   }
    "openapi" {
        docker run --rm -ti -v ${ProjectRoot}:/local openapitools/openapi-generator-cli generate -c /local/scripts/generator-cfg.yaml
    }
    default {
        throw "Unknown command: $command"
    }
}