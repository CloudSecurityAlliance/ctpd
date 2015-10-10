
function MakeId(s) {
    if (s) {
        oid = ObjectId(s)
    } else {
        oid = ObjectId()
    }
    r = HexData(0,oid.str).base64().replace(/\+/g, '-').replace(/\//g, '_')
    print("Creating " + r)
    return r
}


conn = new Mongo();
db = conn.getDB("ctp");

db.dropDatabase()

db = conn.getDB("ctp");

/*********************/
access_id = MakeId()
access_id_tag = "id:" + access_id

x = {
    "_id": access_id,
    "name": "ordinary user",
    "accessTags": [ access_id_tag, "access:user", "access:anybody" ],
    "token": "1234",
}

db.access.insert(x);

x = {
    "_id": MakeId(),
    "name": "super user",
    "accessTags": [ "*" ],
    "token": "0000",
}

db.access.insert(x);

/*********************/
serviceViewId = MakeId("548610513d561badbc7e16d0")

x = {
    "_id": serviceViewId, 
    /* no parent */
    "accessTags": [ access_id_tag ],
    "annotation": "Home network information service",
    "provider": "matrix.lan"
}

db.serviceViews.insert(x);

/*********************/
/* TODO: We should probably index on "key" in real life in the "tokens" collection */


/**********************/

assetId = MakeId("548610513d561badbc7e16d4")

x = {
    "_id": assetId,
    "parent": serviceViewId,
    "accessTags": [ access_id_tag ],
    "annotation": "Webserver running on Linux Ubuntu (apache)",
    "name": "https://paris.matrix.lan/"
}

db.assets.insert(x);

/**********************/

attributeId = MakeId("548610513d561badbc7e16d5")

x = { 
    "_id": attributeId,
    "parent": assetId,
    "accessTags": [ access_id_tag ],
    "annotation": "Availability attribute for demo",
    "name": "availability"
}

db.attributes.insert(x);

/**********************/

metricId = MakeId("548610513d561badbc7e16d6")

x = {
    "_id" : metricId,
    /* no parent */
    "accessTags": [ "access:user" ],
    "annotation" : "",
    "baseMetric" : "https://cloudsecurityalliance.org/ctp/metrics#csa:unix-uptime",
    "measurementParameters" : [ ],
    "resultFormat" : [
    { 
        "name": "1-minute-load-average", 
        "type": "number" 
    },
    { 
        "name": "5-minute-load-average", 
        "type": "number" 
    },
    { 
        "name": "15-minute-load-average", 
        "type": "number" 
    }
    ],
}

db.metrics.insert(x)

/**********************/


measurementId = MakeId("548610513d561badbc7e1600")

x = {
    "_id" : measurementId,
    "parent" : attributeId,
    "accessTags": [ access_id_tag ],
    "annotation" : "",
    "metric" : "@/metrics/" + metricId,
    "result" : {
        "value" : [
            { 
                "1-minute-load-average": 1.5,
                "5-minute-load-average": 1.7841, 
                "15-minute-load-average": 0.89 
            }
        ],
        "dateTime" : ISODate()
    },
    "objective" : {
        "condition" : "value[0]['1-minute-load-average']<5",
        "status" : true
    },
    "userActivated": false,
    "state" : "activated"
}

db.measurements.insert(x);

/**********************/
/** PART II ***********/
/**********************/

attributeId = MakeId()

x = { 
    "_id": attributeId,
    "parent": assetId,
    "accessTags": [ access_id_tag ],
    "annotation": "confidentiality of data in transit with SSL/TLS",
    "name": "confidentiality"
}

db.attributes.insert(x);

/**********************/

metricId = MakeId()

x = {
        "_id" : metricId,
        "accessTags": [ access_id_tag ],
        "annotation" : "",
        "baseMetric" : "https://cloudsecurityalliance.org/ctp/metrics#csa:cryptographic-strength",
        "measurementParameters" : [ 
                {
                    "name": "scale",
                    "type": "string",
                    "value": "ECRYPT II"
                }
            ],
        "resultFormat" : [
            { 
                "name": "level", 
                "type": "number" 
            },
        ],
}

db.metrics.insert(x);

/**********************/

measurementId = MakeId()

    x = {
        "_id" : measurementId,
        "parent" : attributeId,
        "accessTags": [ access_id_tag ],
        "annotation" : "",
        "metric" : "@/metrics/" + metricId,
        //"metric" : metricId,
        "result" : {
            "value" : [
                { 
                    "level": 7,
                }
            ],
            "dateTime" : ISODate()
        },       
        "objective" : {
            "condition" : "value[0].level>=7",
            "status" : true
        },
        "userInitiated" : "false",
        "state" : "activated"

}


db.measurements.insert(x);

/*****************/

print("Dummy schema loaded: OK.")
