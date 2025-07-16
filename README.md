# Excel-Replacement-Website
My workout spreadsheet is starting to chug a little, which means that my programmer brain cannot rest until I make a more optimal solution. I might expand this project to work with my other spreadsheets for diet and finances at some point.

I chose the tech stack for the following reasons:
- Go + HTMX for performance and simplicity reasons, in addition to the fact that I really want to get some practice with these technologies
- SQLite because I'm just going to run the server locally anyways, and I want to be able to load the database from a file to give me the flexibility of, say, loading the db from my onedrive.

## Lessons Learned

I found out that what I'm doing with this app is effectively a CRUD app. It really made me understand why web devs think AI is gonna replace all engineers.

Getting this website "production ready" would require a lot of work. I would need to refactor the ORM to point at an external DB, which would be pretty easy. The harder part would be implementing auth so that a user's data is private and actually associated with them. I've done auth before professionally, I just think it's out of scope for this project.

The use of an ORM means that it would be pretty trivial to switch databases, but overall I think that using an ORM was a mistake because of how much it obfuscates what's actually happening behind the scenes. Things like using a foreign key can just fail silently or certain fields relating one model to another could be blank, and the only way to fix the problem to go diving into the docs to figure out what's happening. GORM in particular seems to be very opinionated about how you should use it, but it won't tell what those opinions are. That isn't even getting into the performance problems of using ORMs vs hand crafted SQL

Using raw sql worked out pretty well for the nutrition database. I had to switch to that because GORM quite simply couldn't batch inserts properly when processing the USDA data. Specifically, the batching function worked, but it would trigger inserts for related data, which would overlaod the query engine. Raw sql is very verbose, especially when processing many to many relationships, but it's manageable, especially with supermaven.

Also, I found out the hard way that sqlite has dynamic typing for individual values ._.