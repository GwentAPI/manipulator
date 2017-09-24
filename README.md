# Manipulator
Manipulator is a CLI tool to help manage GwentAPI

This application is a tool to quickly perform maintenance operation on GwentAPI database and application.

## Update the mongoDB database

When a new release of Gwent happens you will have to generate the json file containing the card definitions in the standard format as defined by *Gwent Community Developers*.

Once you have the file, you can use *manipulator* to update the GwentAPI service.

Before the current database is updated, the tool will create a backup of the databases of your local mongod under ``./backup/``.

``./manipulator update --input <pathToFile.json> --db gwentapi``

If the ``--db`` flag is not specified, the default database of mongoDB will be used: test.

You may use the ``--ssl`` option to connect to the database with ssl. If your database is not accessible from localhost, you can specify different hosts with the ``--host`` flag.

``./manipulator update --input <pathToFile.json> --db gwentapi --host <myAddress>``

If you are running a replica set, you can supply a list of hosts like so:

```
./manipulator update --input <pathToFile.json> --db gwentapi --host \
"host1[:porthost1],
host2[:porthost2],
host3[:porthost3]"
```

## Download the new artworks

As per the design of the standard format, card artworks are available from an URI. To download the new artworks, run the following command:

``./manipulator artwork --input <pathToFile.json>``

By default, the artwork will be saved under the ``./artworks/`` directory. You can specify a different path by using the ``--out`` flag:

``./manipulator artwork --input <pathToFile.json> --out <outputPath>``

## Backup the database

You can backup the databases of your local mongod without being in the process of updating the db:

``./manipulator backup``

Or for a remote database (replica set supported):

``./manipulator backup --u username --p password --authenticationDatabase admin --ssl --host "host1,host2,host3"``

## Additional help

You can run the ``--help`` flag on the program or on specific commands to learn more.

## WIP and planned features

* Being able to specify backup destination.
* Better backup archive structure to add contextual info like:
    
    * Which server was updated?
    * What was the card definition version?
* Rollback feature.