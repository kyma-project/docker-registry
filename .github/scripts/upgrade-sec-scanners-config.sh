#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

yq eval-all --inplace '
    select(fileIndex == 0).bdba=[
        select(fileIndex == 1)
        | .containerRegistry.path as $registryPath
        | (
            .images + {
                "dockerregistry_operator" : {
                    "name" : "dockerregistry-operator",
                    "directory" : "prod",
                    "version" : env(IMG_VERSION)
                }
            }
          )[]
        | $registryPath + "/" + .directory + "/" + .name + ":" + .version
    ]
    | select(fileIndex == 0)
    ' sec-scanners-config.yaml config/docker-registry/values.yaml