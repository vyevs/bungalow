# Dev Environment Setup

This directory provides all the infrastructure, in the form of a docker compose file. `Taskfile.dist.yml` defines commands to make working with the infrastructure easy.

If you'd like to modify the taskfile, don't modify the `.dist` version, create your own and do whatever you like with it:
```
cp Taskfile.dist.yml Taskfile.yml`
```
Note that `Taskfile.yml` is git-ignored. This allows for individuals to have their own custom taskfiles without any conflicts.

If you feel there is a command that would be useful to have in the `.dist` taskfile, add it.

## Bring up the infrastructure
```
task up
```
By default, the containers will run in attached mode, connected to your terminal. This means you will see container logs. `Ctrl+C` will stop and remove all containers.

## Bring down the infrastructure

When using `task up` in attached mode, you don't need to explicitly bring down the infrastructure. Simply `Ctrl+C` and the containers will be stopped and removed.

You can use `task down` to stop and remove any detached containers.

## Postgres

The postgres container can be accessed at `localhost:5432`.
The admin user is `bungalow` and password is `password`.

## PGAdmin

A PGAdmin container that exposes a web interface is available on `localhost:5050`.

Log in using email `bungalow@email.com` and password `password`.


## Persistence

All our containers persist their data using bind mounts. Data is persisted to directories in the `data` directory. e.g. postgres data is in `data/postgres`, prometheus data is in `data/prometheus`.

The postgres and prometheus data directories are git-ignored so you have your own personal copy. 

Grafana and pgadmin data are not git-ignored, because dashboards and pgadmin setup is common to everyone.