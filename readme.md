# Random Notion

This is a small sample project for playing around with the Notion API.
It's based on [Tiago Forte's idea](https://fortelabs.co/blog/p-a-r-a-iii-building-an-idea-generator/)
of a utility to fetch random pages you've previously stored in EverNote.
This project expects a central "content" database in which all pages are stored.
It queries the Notion API to build a list of all pages in the database, then
returns one at random, which when opened should redirect to Notion to view.

## Caveats

This is just a personal project right now, with no intention of allowing others to
connect it to their accounts. My implementation of a central database for content is
rather specific, so I'm not sure whether this could be made more generally useful.

## Possible improvements

Depending on how far I decide to take this, possible improvements include:

- Cache retrieved pages to improve performance (right now, each invocation retrieves all pages)
- Implement Notion auth to enable others to add this as an integration
  - This would probably also require making this less tied to my specific Notion structure