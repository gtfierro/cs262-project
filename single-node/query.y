//# vi:syntax=go
%{
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/taylorchu/toki"
	"gopkg.in/mgo.v2/bson"
)

// AST structures

type key string
type value string
type Node interface{
    // generates the mongo query for this node
    MongoQuery() bson.M
}

type equalNode struct {
	Key		string
	Value	string
}

func (en equalNode) MongoQuery() bson.M {
    return bson.M{en.Key: en.Value}
}

type nequalNode struct {
	Key		string
	Value	string
}

func (nen nequalNode) MongoQuery() bson.M {
    return bson.M{nen.Key: bson.M{"$neq": nen.Value}}
}

type likeNode struct {
	Key	string
	Value	string
}

func (ln likeNode) MongoQuery() bson.M {
    return bson.M{ln.Key: bson.M{"$regex": ln.Value}}
}

type hasNode struct {
	Key string
}

func (hn hasNode) MongoQuery() bson.M {
    return bson.M{hn.Key: bson.M{"$exists": true}}
}

type andNode struct {
	Left	Node
	Right	Node
}

func (an andNode) MongoQuery() bson.M {
    return bson.M{"$and": []bson.M{an.Left.MongoQuery(), an.Right.MongoQuery()}}
}

type orNode struct {
	Left	Node
	Right	Node
}

func (on orNode) MongoQuery() bson.M {
    return bson.M{"$or": []bson.M{on.Left.MongoQuery(), on.Right.MongoQuery()}}
}

type notNode struct {
	Clause	Node
}

//TODO: fix this!
func (nn notNode) MongoQuery() bson.M {
    return nn.Clause.MongoQuery()
}

type rootNode struct {
	Keys	[]string
	Tree	Node
	String	string
}

//TODO: root node is an interface with "getkeys" method

%}

%union{
    str     string
    node    Node
}

%token <str> LPAREN RPAREN AND OR NOT LIKE EQ NEQ HAS
%token <str> KEY VALUE

%type <node> predicate term
%type <str> key value

%right EQ

%%
root   :    predicate
            {
                cqbslex.(*Lexer).root = rootNode{Keys: keys, Tree: $1}
            }
            ;

predicate   :  predicate AND term
            {
                $$ = andNode{Left: $1, Right: $3}
            }
            |  predicate OR term
            {
                $$ = orNode{Left: $1, Right: $3}
            }
            |  NOT term
            {
                $$ = notNode{Clause: $2}
            }
            |  term
            {
                $$ = $1
            }
            ;

term        :  key LIKE value
            {
                $$ = likeNode{Key: $1, Value: $3}
            }
            |  key EQ value
            {
                $$ = equalNode{Key: $1, Value: $3}
            }
            |  key NEQ value
            {
                $$ = nequalNode{Key: $1, Value: $3}
            }
            |  HAS key
            {
                $$ = hasNode{Key: $2}
            }
            |  LPAREN predicate RPAREN
            {
                $$ = $2
            }
            ;

key         : KEY
            {
                keys = append(keys, $1)
                $$ = $1
            }
            ;

value       : VALUE
            {
              $$ = $1[1:len($1)-1]
            }
%%

const eof = 0
var keys = []string{}

func Parse(querystring string) rootNode {
    lexer := &Lexer{query: querystring}
    lexer.scanner = toki.NewScanner(
        []toki.Def{
            {Token: AND, Pattern: "and"},
            {Token: OR, Pattern: "or"},
            {Token: NOT, Pattern: "not"},
            {Token: HAS, Pattern: "has"},
            {Token: NEQ, Pattern: "!="},
            {Token: EQ, Pattern: "="},
            {Token: LIKE, Pattern: "like"},
            {Token: LPAREN, Pattern: "\\("},
            {Token: RPAREN, Pattern: "\\)"},
            {Token: KEY, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%\\_\\-]*"},
			//{Token: KEY, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%_\\-]*"},
			{Token: VALUE, Pattern: "(\"[^\"\\\\]*?(\\.[^\"\\\\]*?)*?\")|('[^'\\\\]*?(\\.[^'\\\\]*?)*?')"},
    })
	lexer.scanner.SetInput(querystring)
    cqbsParse(lexer)
    root := lexer.root
    lexer.keys = keys
    log.WithFields(log.Fields{
        "string": querystring, "keys": keys, "query": lexer.query, "root": root,
    }).Info("Finished parsing query")
    keys = []string{}
	root.String = lexer.query
    return root
}

type Lexer struct {
    query   string
    scanner *toki.Scanner
    keys    []string
    node    *Node
    root    rootNode
    lasttoken string
}

func (l *Lexer) Lex(lval *cqbsSymType) int {
	r := l.scanner.Next()
    l.lasttoken = r.String()
	if r.Pos.Line == 2 || len(r.Value) == 0 {
		return eof
	}
    lval.str = string(r.Value)
	return int(r.Token)
}

func (l *Lexer) Error(s string) {
    log.WithFields(log.Fields{
        "token": l.lasttoken, "query": l.query, "error": s,
    }).Error("Error parsing query")
}
