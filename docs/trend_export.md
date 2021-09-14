# Trend Export Tool

The trend_export tool retrieves trends from one or more WebCTRL
servers and exports them as a tab-separated (TSV) file suitable
for data analysis or visualization in many tools including (but
by no means limited to) Microsoft Excel.  Trend data is obtained
using the WebCTRL SOAP interface, so nothing has to be installed
on the WebCTRL server.

Internally, trend_export requests a few thousand points at a time
for each trend then merges them together using a streaming merge
sort.  This helps minimize performance impact on the server, even
if you request years of data at once, as the data will be pulled
in chunks then stitched together.

This tool requires no external dependencies.  You do not have to
install python, java, or anything else.  There is no installer,
and it runs on most popular platform including Linux, Winmdows,
and Darwin (Apple).  This tool is NOT open source today but
may become so some time in the future.

**Security Note: SOAP is a separate operator privilege in WebCTRL.  For security reasons,
we recommend you create a separate SOAP user with read-only access.
Under no circumstances should you be using the administrator login
to make remote requests.**


## Usage

-user _username_ - The WebCTRL login used to make trend requests.  This
user must have SOAP privileges as well as read privileges to for the
locations being requested (if location-based permissions are enabled
in the server).  If no username is specified either on the command
line, in a credentials file, or in a request configuration file, then
the username will be prompted.

-password _password_ - The password for the WebCTRL login used to make
trend requests.   If no password is specifeid either on the command
line, in a credentials file, or in a request configuration file, then
the password will be prompted.  Note that it is generally not a good
idea to specify a password on the command line as it may be visible
through _/proc_ or _ps_.

-creds _path_ - The username and password are stored in an external
file.  The format of this file is text, exactly one line long,
with the username and password separated by a colon, for example:

    spongebob:squarepants

-server _url_ - the base URL to the WebCTRL server.  It is usually
something like _https://some.server.com:1443/_ which may include the
port number (if not, 443 will be assumed for https)

-start _YYYY-MM-DDTHH:MM:SS_ the starting time for the request.  Note that due
to the behavior of WebCTRL the first point returned will be the
first point _after_ the start time, so if you need the start time
included start your request one second earlier.  For format
see the time handling section below.

-stop _YYYY-MM-DDTHH:MM:SS_ the starting time for the request. For format
see the time handling section below.

-config _configfile_ - specify the list of servers and trends in
an external file (see below for format)

## Time and Date handling

WebCTRL time and date handling can be problematic via the SOAP
interface, mainly because GetTrend() expects and returns dates as
strings.  The WebCTRL date format does not include a time zone,
so requests must be made using whatever time zone the server
uses.

This tool accepts start and stop dates in this exact format:

    YYYY-MM-DDTHH:MM:SS

for example:

    trend_export -start 2021-01-01T15:00:00 -stop 2021-02-01T00:00:00 ...

In the output, the time format is the same but without the T:

    "time"                  "zone_temp"
    "2021-09-13 12:15:00"	72.099998
    "2021-09-13 12:30:00"	71.800003

Trend exports will be affected by the transition from daylight
savings to standard time and back if the server is in a time
zone that supports daylight savings time.  Unfortunately, this
is a problem with the WebCTRL SOAP interface and there isn't an
easy way around it.  The effect is that you will see a missing hour
during the transition to standard time (spring ahead), and an
overlapping hour in the transition to daylight savings (fall back).
When using point interpolation (the default) this will make it look
like your trends all got stuck for an hour  in the spring.  The
other change will not be as evident - the WebCTRL server will
return two points for the same datetime, and whichever one the
server returns last will be kept (you should not see duplicate times).

It is hoped that in some future release WebCTRL will have the
option of returning trend timestamps either in epoch offset
(unix time), using a standard time format that reflects the
offset (e.g. ISO8601), or, xsd:dateTime (if the interface remains XML
and SOAP long term).

## External Configuration File

An external configuration file is optional but may be helpful in
certain situations:

  * There are many trends, making a cumbersome, long command line
  * The trends come from different WebCTRL servers
  * You want to give the trends friendlier names in the exported file

The configuration file uses the YAML format.  In YAML, spacing is used
for grouping.  It doesn't matter if you use spaces or tabs, it only
matters that you be consistent.

The configuration file consists of a list of servers each with a
list of trend locations for each server.  Normally there would only
be a single server line unless you are a complex site with multiple,
hierarchical WebCTRL servers.

You can specify a login and password in this file if you choose,
but it is optional.  If you don't specify a login and password
somewhere you will be prompted.  This is probably preferable
for manual extracts, but the ability to specify a login and
password in the config file may be helpful if running on a
schedule (e.g. through cron or the windows job scheduler).

An example configuration file is:

```yaml
servers:
  - url: "https://server1:442/"
    login: "spongebob"
    password: "squarepants"
    locations:
      - location: "/#room1234/trend_name"
        name: "Room 1234 zone temp"
      - location: "/#vav_1234/chw_pct_trn"
        name: "Room 1234 CHW valve %"
  - url: "https://server2:442/"
    login: "patrick"
    password: "star"
    locations:
      - location: "/#room2345/trend_name"
        name: "Room 2345 zone temp"
      - location: "/#vav_2345/chw_pct_trn"
  -       name: "Room 2345 CHW valve %"

```
