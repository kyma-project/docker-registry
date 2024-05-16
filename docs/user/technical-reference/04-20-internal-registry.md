# Internal Docker Registry

The Docker Registry module comes with the internal Docker Registry, which stores the container images.

The internal Docker Registry is not recommended for production, as it's not deployed in the High Availability (HA) setup and has limited storage space and no garbage collection of the orphaned images.

Still, it is very convenient for development and getting first-time experience with the Kyma Docker Registry.
