CurateND Cookbook
=================

This file answers questions about the CurateND batch ingest
and how to use ROF to accomplish specific tasks.

# Making Old-Style Collections in CurateND

The only way to create an old style collection is to create an ROF file
describing it, and then run the file through the batch ingest.

The base collection object follows. Copy it and then update the `rights`,
`properties`, `dc:title`, `dc:description`, `content-file`, and
`thumbnail-file`. The content and thumbnail datastreams is an image associated
with the collection: the content is the large version and the thumbnail is the
small.

    {
      "type": "fobject",
      "af-model": "Collection",
      "rights": {
        "read-groups": [
          "public"
        ],
        "edit": [
          "dbrower"
        ]
      },
      "properties": "<fields>\n<depositor>dbrower</depositor>\n<owner>dbrower</owner>\n</fields>\n",
      "properties-meta": {
        "mime-type": "text/xml"
      },
      "metadata": {
        "dc:title": "Put The Collection Title Here",
        "dc:description": "Put the description here",
        "@context": {
          "dc": "http://purl.org/dc/terms/",
          "foaf": "http://xmlns.com/foaf/0.1/",
          "rdfs": "http://www.w3.org/2000/01/rdf-schema#",
          "dc:dateSubmitted": {
            "@type": "http://www.w3.org/2001/XMLSchema#date"
          },
          "dc:modified": {
            "@type": "http://www.w3.org/2001/XMLSchema#date"
          }
        }
      },
      "content-file": "image.png",
      "content-meta": {
        "mime-type": "image/png"
      },
      "thumbnail-file": "image-thumb.png",
      "thumbnail-meta": {
        "mime-type": "image/png"
      }
    }

# Adding Items To A Collection

There are two ways to add a work to a collection. If the work and collection already exist in
Fedora, you can make a new batch ingest job containing only the files `metadata-1.collection` and `JOB`.
The file `metadata-1.collection` should be a JSON object where each key is a collection PID, and its
value is a list of work PIDs to be added to it. For example:

    {
        "und:collection1": [
            "und:work1",
            "und:work2",
            "und:work3"
        ]
    }

And the `JOB` file should look like the following:

    {
        "Todo": ["submit-collections"]
    }

The other way to add a work to a collection is to list the collections it should
belong to in its ROF under the "collections" label. For example:

    {
        "type": "fobject",
        "af-model": "Work",
        "collections": [
            "und:collection1",
            "und:collection2"
        ]
    }

This work will be given a PID and then added to the collections represented by
`und:collection1` and `und:collection2`.

# Updating Item Metadata (Brute Force)

One way to update an item's metadata is to alter the usual `metadata` section
and then submit it for batch ingest. This way is preferable for items where a
`metadata` in JSON-LD is already available. Another way is to update the
ntriples stored for an item directly, which is much more expiedent (at least for
now, 2016-04-30). Here is how to do it.

First, download the ntriples of the item you want to update, say noid
`1234567890b`. Replace the URL with the approprate one to reach the fedora
instance the item is stored in.

    curl https://fedora.url:0000/objects/und:1234567890b/datastreams/descMetadata/content --user fedoraAdmin:fedoraAdmin > 1234567890b.nt

There should be a file `1234567890b.nt` contains the metadata in ntriples
format. Edit the ntriples file directly (and update the `dc:modified` term if
you are feeling generous), and then put into a directory. We need to create the
ROF file, and that is a little complicated since we need to provide an
ActiveFedora model and the RELS-EXT of the original item. (This is not
straightforward, this section should be expanded. Sorry in advance for not doing
that.) The final `metadata-1.rof` should look similar to the following:

    [{
        "pid": "1234567890b",
        "af-model": "Etd",
        "rels-ext":{
            ... (TODO: fill in)
        },
        "descMetadata-meta" : {
            "mimetype": "text/plain"
        },
        "descMetadata-file": "1234567890b.nt"
    }]

The important lines are the last four, which tell the batch ingester to replace the
`descMetadata` datastream with the new ntriples file we have.

Now create a `JOB` file in the directory with the following contents:

    {
        "Todo": ["verify", "ingest", "index"]
    }

The special `JOB` makes sure the date stamping step of the ingest process is not
run. Submit the job, and the metadata will now be updated.

(Caveat: this will need to be modified slightly for bendo...the bendo-item will
need to be pulled from fedora and stuck into the rof file.)
