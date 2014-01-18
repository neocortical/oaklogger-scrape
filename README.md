oaklogger-scrape
================

Code for scraping public posts in Oaklog.com and building a full-text search utility in MongoDB.

Dependency: 
* MongoDB v 2.4.x

MongoDB setup: 

Four collections are required: users, threads, posts, and status.

Indexes to set up:

status:
db.status.ensureIndex({"pid":1},{unique:true});

threads:
db.threads.ensureIndex({"tid":1}, {unique:true});
db.threads.ensureIndex( { "name": "text" }, {unique : true} );

posts:
db.posts.ensureIndex( { "message": "text" } );
db.posts.ensureIndex({"tid":1});
db.posts.ensureIndex({"pid":1},{unique:true});

users:
db.users.ensureIndex({ uid:1 }, { unique:true });

