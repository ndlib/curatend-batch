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
