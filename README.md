Cloud Trust Protocol Daemon Prototype
=====================================

This repository contains a prototype server implementing the Cloud Security
Alliance's [Cloud Trust Protocol](https://cloudsecurityalliance.org/group/cloudtrust-protocol/).

:warning: For the latest release of the __ctpd__ source code please always
refer to: https://github.com/cloudsecurityalliance/ctpd

This prototype called __ctpd__ is a unix-style server written in
[Go](http://golang.org) with mongodb as a database backend. It has been tested
on Ubundtu/Debian Linux and Mac OS X. The code of __ctpd__ is still in 'beta'
stage and is mainly intended for testing and research purposes.


The __ctpd__ server provided in this repository:

* Implements most the CTP API defined in the [CTP Data model and API specification][CTP] [(pdf)][CTP_pdf] defined by Cloud Security Alliance.

* Implements a ['back office' API][CTPBO] that allows to update the database managed by ctpd. __This extra API is not
part of the official CTP specification__.

Trying ctpd
-----------

To compile and run **ctpd**, you will need to install:

* go (http://golang.org),
* mongodb (https://www.mongodb.org/) and

The ctpd source code is expected to reside in
`$GOPATH/src/github.com/cloudsecurityalliance/ctpd`. 

The easiest way to download the ctpd source code for a test drive is simply to
type:


    go get github.com/cloudsecurityalliance/ctpd  


For alternative ways to install ctpd, see the INSTALL file.

Next, run the mongodb script `build_db.js` in the `tools/` subdirectory of this
source code repository:


    mongo build_db.js


Launch __ctpd__ for a test drive as follows:


    go run ctpd.go


By default, __ctpd__ runs on port 8080, and you can test that is working with a
simple curl command:


    curl -H "Authorization: Bearer 1234" http://localhost:8080/api/1.0/


The result should look something like this:

```json
{
  "self": "http://localhost:8080/api/1.0/",
  "name": "",
  "annotation": "ctpd prototype server",
  "version": "0",
  "provider": "",
  "serviceViews": "http://localhost:8080/api/1.0/serviceViews",
  "metrics": "http://localhost:8080/api/1.0/metrics"
}
```

Note that the value "1234" above is not an example of a secure token and was
created by `build_db.js` for demonstration purposes only.

You can also test the embedded lightweight javscript client by launching ctpd
as follows:


    go run ctpd.go --color-logs --client=./client/


Then point your browser to http://ctpserver:8080/ where 'ctpserver' should be
replaced by the hostname of the machine that is running ctpd.

Using ctpd
----------

_More information should be provided here soon._

Specification coverage
----------------------

**ctpd** aims to fully implement the [the CTP data model and API][CTP] [(pdf)][CTP_pdf], as well as the non official
[CTP 'back office' API][CTPBO]. 

The following table summarizes the level of implementation of the CTP data model and API specification in **ctpd**, as of December 2015:

Specification                  | Implementation status in prototype
-------------------------------|-----------------------------------------------
Service views                  | **100%**
Assets                         | **100%**
Attributes                     | **100%**
Measurements                   | **100%**
Triggers                       | **50%** (_missing trigger deletion_)
Logs                           | **100%**
Dependencies                   | _0%_
XMPP notification              | _0%_
CTPScript interpreter          | _90%_
SSL/TLS (as an option)         | **100%**
OAuth Bearer token auth.       | **100%**

The following table summarizes the level of implementation of the CTP 'back office' API in **ctpd**:

Specification                     | Implementation status in prototype
----------------------------------|-----------------------------------------------
Resource creation                 | **95%**: (_missing for dependencies_)
Resource deletion                 | **90%**: (_missing for dependencies and logs_)
Resource access control with tags | **100%**
Account creation                  | **100%**
Account deletion                  | **100%**
Account modification              | _0%_
XMPP backend                      | _0%_
Embedded javascript client option | **90%** (_missing configuration of entry point_)

:warning: No formal code security analysis has been conducted yet.

About CTP
---------
The Cloud Trust Protocol (CTP) is designed to be a mechanism by which cloud
service customers can ask for and receive information related to the security
of the services they use in the cloud, promoting transparency and trust.

The Cloud Security Alliance has defined a "CTP data model and API", which
specifies how monitoring information should be presented to cloud customers.
For more details, see [the data model and API][CTP] [(pdf)][CTP_pdf].

The actual implementation of the API is left to the choice of the service
provider, and the source code provided here is just an example of such an
implementation, without any normative value.

Licence
-------

ctpd is copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)
and is licensed under the Apache License, Version 2.0, as described
in the LICENCE file.

ctpd contains an optional demo client which uses:

* bootstrap, licenced under the MIT licence
as described in [here](https://github.com/twbs/bootstrap/blob/master/LICENSE).

* jstree, licenced under the MIT licence
as described in [here](https://github.com/vakata/jstree/#license--contributing).

[CTP]: http://htmlpreview.github.io/?https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.html
[CTP_pdf]: https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.pdf
[CTPBO]: http://htmlpreview.github.io/?https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Admin-API.html
