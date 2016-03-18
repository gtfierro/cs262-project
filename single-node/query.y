//# vi:syntax=go
%{
package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/taylorchu/toki"
)

// AST structures

type key string
type value string
type node interface{}

type equalNode struct {
	Key		key
	Value	value
}

type nequalNode struct {
	Key		key
	Value	value
}

type likeNode struct {
	Key	key
	Value	value
}

type hasNode struct {
	Key key
}

type andNode struct {
	Left	node
	Right	node
}

type orNode struct {
	Left	node
	Right	node
}

type notNode struct {
	Clause	node
}

var root node
%}

%union{
    str     string
    key     key
    value   value
    node    node
}

%token <str> LPAREN RPAREN AND OR NOT LIKE EQ NEQ HAS
%token <key> KEY
%token <value> VALUE

%type <node> predicate term
%type <key> key

%right EQ

%%
root   :    predicate
            {
                root = $1
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

term        :  key LIKE VALUE
            {
                $$ = likeNode{Key: $1, Value: $3}
            }
            |  key EQ VALUE
            {
                $$ = equalNode{Key: $1, Value: $3}
            }
            |  key NEQ VALUE
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
%%

const eof = 0
var keys = []key{}

func Parse(querystring string) {
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
    lexer.keys = keys
    log.WithFields(logrus.Fields{
        "keys": keys, "query": lexer.query,
    }).Info("Finished parsing query")
    log.Debug(root)
}

type Lexer struct {
    query   string
    scanner *toki.Scanner
    keys    []key
    node    *node
    lasttoken string
}

func (l *Lexer) Lex(lval *cqbsSymType) int {
	r := l.scanner.Next()
    l.lasttoken = r.String()
	if r.Pos.Line == 2 || len(r.Value) == 0 {
		return eof
	}
    switch r.Token {
    case KEY:
        lval.key = key(r.Value)
    case VALUE:
        lval.value = value(r.Value)
    default:
        lval.str = string(r.Value)
    }
	return int(r.Token)
}

func (l *Lexer) Error(s string) {
    log.WithFields(logrus.Fields{
        "token": l.lasttoken, "query": l.query, "error": s, 
    }).Error("Error parsing query")
}
