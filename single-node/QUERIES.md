Query Language
==============

The [Giles/sMAP query language](https://gtfierro.github.io/giles/interface/#querylang) has
a flexible construction and should cover most of the use cases for now. There is need for
a discussion on whether CQBS permits subscribing to SELECT clauses. For now, we will implement
subscribing to WHERE clauses.

## Syntax

Queries are strings consisting of the following binary/unary operators.

| Operator | Description | Usage | Example |
|:--------:| ----------- | ----- | ------  |
|  `=`     | Compare tag values.  | `tagname = "tagval"` | `Metadata/Location/Building = "Soda Hall"` |
| `like`  | String matching. Use Perl-style regex | `tagname like "pattern"` | `Metadata/Instrument/Manufacturer like "Dent.*"` |
| `has`    | Filters streams that have the provided tag | `has tagname` | `has Metadata/System` |
| `and`    | Logical AND of two queries (on either side) | `where-clause and where-clause` | `has Metadata/System and Properties/UnitofTime = "s"` |
| `or`     | Logical OR of two queries | | |
| `not`    | Inverts a where clause | `not where-clause` | `not Properties/UnitofMeasure = "volts"` |
| ON HOLD: `in`     | Matches set intersection on lists of tags | `[list,of,tags] in tagname` | `["zone","temp"] in Metadata/HaystackTags` | 
