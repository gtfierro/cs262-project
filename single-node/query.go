//line query.y:3
package main

import __yyfmt__ "fmt"

//line query.y:3
import (
	log "github.com/Sirupsen/logrus"
	"github.com/taylorchu/toki"
	"gopkg.in/mgo.v2/bson"
)

// AST structures

type key string
type value string
type Node interface {
	// generates the mongo query for this node
	MongoQuery() bson.M
}

type equalNode struct {
	Key   string
	Value string
}

func (en equalNode) MongoQuery() bson.M {
	return bson.M{en.Key: en.Value}
}

type nequalNode struct {
	Key   string
	Value string
}

func (nen nequalNode) MongoQuery() bson.M {
	return bson.M{nen.Key: bson.M{"$ne": nen.Value}}
}

type likeNode struct {
	Key   string
	Value string
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
	Left  Node
	Right Node
}

func (an andNode) MongoQuery() bson.M {
	return bson.M{"$and": []bson.M{an.Left.MongoQuery(), an.Right.MongoQuery()}}
}

type orNode struct {
	Left  Node
	Right Node
}

func (on orNode) MongoQuery() bson.M {
	return bson.M{"$or": []bson.M{on.Left.MongoQuery(), on.Right.MongoQuery()}}
}

type notNode struct {
	Clause Node
}

//TODO: fix this!
func (nn notNode) MongoQuery() bson.M {
	return nn.Clause.MongoQuery()
}

type rootNode struct {
	Keys   []string
	Tree   Node
	String string
}

//TODO: root node is an interface with "getkeys" method

//line query.y:92
type cqbsSymType struct {
	yys  int
	str  string
	node Node
}

const LPAREN = 57346
const RPAREN = 57347
const AND = 57348
const OR = 57349
const NOT = 57350
const LIKE = 57351
const EQ = 57352
const NEQ = 57353
const HAS = 57354
const KEY = 57355
const VALUE = 57356

var cqbsToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"LPAREN",
	"RPAREN",
	"AND",
	"OR",
	"NOT",
	"LIKE",
	"EQ",
	"NEQ",
	"HAS",
	"KEY",
	"VALUE",
}
var cqbsStatenames = [...]string{}

const cqbsEofCode = 1
const cqbsErrCode = 2
const cqbsInitialStackSize = 16

//line query.y:163

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
	query     string
	scanner   *toki.Scanner
	keys      []string
	node      *Node
	root      rootNode
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

//line yacctab:1
var cqbsExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const cqbsNprod = 13
const cqbsPrivate = 57344

var cqbsTokenNames []string
var cqbsStates []string

const cqbsLast = 32

var cqbsAct = [...]int{

	7, 19, 7, 20, 3, 8, 9, 10, 6, 8,
	6, 8, 12, 13, 14, 21, 22, 4, 23, 9,
	10, 11, 2, 1, 5, 0, 0, 17, 18, 0,
	16, 15,
}
var cqbsPact = [...]int{

	-4, -1000, 0, -2, -1000, 3, -8, -4, -1000, -2,
	-2, -1000, -11, -11, -11, -1000, 13, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000,
}
var cqbsPgo = [...]int{

	0, 22, 17, 24, 1, 23,
}
var cqbsR1 = [...]int{

	0, 5, 1, 1, 1, 1, 2, 2, 2, 2,
	2, 3, 4,
}
var cqbsR2 = [...]int{

	0, 1, 3, 3, 2, 1, 3, 3, 3, 2,
	3, 1, 1,
}
var cqbsChk = [...]int{

	-1000, -5, -1, 8, -2, -3, 12, 4, 13, 6,
	7, -2, 9, 10, 11, -3, -1, -2, -2, -4,
	14, -4, -4, 5,
}
var cqbsDef = [...]int{

	0, -2, 1, 0, 5, 0, 0, 0, 11, 0,
	0, 4, 0, 0, 0, 9, 0, 2, 3, 6,
	12, 7, 8, 10,
}
var cqbsTok1 = [...]int{

	1,
}
var cqbsTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14,
}
var cqbsTok3 = [...]int{
	0,
}

var cqbsErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	cqbsDebug        = 0
	cqbsErrorVerbose = false
)

type cqbsLexer interface {
	Lex(lval *cqbsSymType) int
	Error(s string)
}

type cqbsParser interface {
	Parse(cqbsLexer) int
	Lookahead() int
}

type cqbsParserImpl struct {
	lval  cqbsSymType
	stack [cqbsInitialStackSize]cqbsSymType
	char  int
}

func (p *cqbsParserImpl) Lookahead() int {
	return p.char
}

func cqbsNewParser() cqbsParser {
	return &cqbsParserImpl{}
}

const cqbsFlag = -1000

func cqbsTokname(c int) string {
	if c >= 1 && c-1 < len(cqbsToknames) {
		if cqbsToknames[c-1] != "" {
			return cqbsToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func cqbsStatname(s int) string {
	if s >= 0 && s < len(cqbsStatenames) {
		if cqbsStatenames[s] != "" {
			return cqbsStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func cqbsErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !cqbsErrorVerbose {
		return "syntax error"
	}

	for _, e := range cqbsErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + cqbsTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := cqbsPact[state]
	for tok := TOKSTART; tok-1 < len(cqbsToknames); tok++ {
		if n := base + tok; n >= 0 && n < cqbsLast && cqbsChk[cqbsAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if cqbsDef[state] == -2 {
		i := 0
		for cqbsExca[i] != -1 || cqbsExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; cqbsExca[i] >= 0; i += 2 {
			tok := cqbsExca[i]
			if tok < TOKSTART || cqbsExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if cqbsExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += cqbsTokname(tok)
	}
	return res
}

func cqbslex1(lex cqbsLexer, lval *cqbsSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = cqbsTok1[0]
		goto out
	}
	if char < len(cqbsTok1) {
		token = cqbsTok1[char]
		goto out
	}
	if char >= cqbsPrivate {
		if char < cqbsPrivate+len(cqbsTok2) {
			token = cqbsTok2[char-cqbsPrivate]
			goto out
		}
	}
	for i := 0; i < len(cqbsTok3); i += 2 {
		token = cqbsTok3[i+0]
		if token == char {
			token = cqbsTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = cqbsTok2[1] /* unknown char */
	}
	if cqbsDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", cqbsTokname(token), uint(char))
	}
	return char, token
}

func cqbsParse(cqbslex cqbsLexer) int {
	return cqbsNewParser().Parse(cqbslex)
}

func (cqbsrcvr *cqbsParserImpl) Parse(cqbslex cqbsLexer) int {
	var cqbsn int
	var cqbsVAL cqbsSymType
	var cqbsDollar []cqbsSymType
	_ = cqbsDollar // silence set and not used
	cqbsS := cqbsrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	cqbsstate := 0
	cqbsrcvr.char = -1
	cqbstoken := -1 // cqbsrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		cqbsstate = -1
		cqbsrcvr.char = -1
		cqbstoken = -1
	}()
	cqbsp := -1
	goto cqbsstack

ret0:
	return 0

ret1:
	return 1

cqbsstack:
	/* put a state and value onto the stack */
	if cqbsDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", cqbsTokname(cqbstoken), cqbsStatname(cqbsstate))
	}

	cqbsp++
	if cqbsp >= len(cqbsS) {
		nyys := make([]cqbsSymType, len(cqbsS)*2)
		copy(nyys, cqbsS)
		cqbsS = nyys
	}
	cqbsS[cqbsp] = cqbsVAL
	cqbsS[cqbsp].yys = cqbsstate

cqbsnewstate:
	cqbsn = cqbsPact[cqbsstate]
	if cqbsn <= cqbsFlag {
		goto cqbsdefault /* simple state */
	}
	if cqbsrcvr.char < 0 {
		cqbsrcvr.char, cqbstoken = cqbslex1(cqbslex, &cqbsrcvr.lval)
	}
	cqbsn += cqbstoken
	if cqbsn < 0 || cqbsn >= cqbsLast {
		goto cqbsdefault
	}
	cqbsn = cqbsAct[cqbsn]
	if cqbsChk[cqbsn] == cqbstoken { /* valid shift */
		cqbsrcvr.char = -1
		cqbstoken = -1
		cqbsVAL = cqbsrcvr.lval
		cqbsstate = cqbsn
		if Errflag > 0 {
			Errflag--
		}
		goto cqbsstack
	}

cqbsdefault:
	/* default state action */
	cqbsn = cqbsDef[cqbsstate]
	if cqbsn == -2 {
		if cqbsrcvr.char < 0 {
			cqbsrcvr.char, cqbstoken = cqbslex1(cqbslex, &cqbsrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if cqbsExca[xi+0] == -1 && cqbsExca[xi+1] == cqbsstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			cqbsn = cqbsExca[xi+0]
			if cqbsn < 0 || cqbsn == cqbstoken {
				break
			}
		}
		cqbsn = cqbsExca[xi+1]
		if cqbsn < 0 {
			goto ret0
		}
	}
	if cqbsn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			cqbslex.Error(cqbsErrorMessage(cqbsstate, cqbstoken))
			Nerrs++
			if cqbsDebug >= 1 {
				__yyfmt__.Printf("%s", cqbsStatname(cqbsstate))
				__yyfmt__.Printf(" saw %s\n", cqbsTokname(cqbstoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for cqbsp >= 0 {
				cqbsn = cqbsPact[cqbsS[cqbsp].yys] + cqbsErrCode
				if cqbsn >= 0 && cqbsn < cqbsLast {
					cqbsstate = cqbsAct[cqbsn] /* simulate a shift of "error" */
					if cqbsChk[cqbsstate] == cqbsErrCode {
						goto cqbsstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if cqbsDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", cqbsS[cqbsp].yys)
				}
				cqbsp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if cqbsDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", cqbsTokname(cqbstoken))
			}
			if cqbstoken == cqbsEofCode {
				goto ret1
			}
			cqbsrcvr.char = -1
			cqbstoken = -1
			goto cqbsnewstate /* try again in the same state */
		}
	}

	/* reduction by production cqbsn */
	if cqbsDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", cqbsn, cqbsStatname(cqbsstate))
	}

	cqbsnt := cqbsn
	cqbspt := cqbsp
	_ = cqbspt // guard against "declared and not used"

	cqbsp -= cqbsR2[cqbsn]
	// cqbsp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if cqbsp+1 >= len(cqbsS) {
		nyys := make([]cqbsSymType, len(cqbsS)*2)
		copy(nyys, cqbsS)
		cqbsS = nyys
	}
	cqbsVAL = cqbsS[cqbsp+1]

	/* consult goto table to find next state */
	cqbsn = cqbsR1[cqbsn]
	cqbsg := cqbsPgo[cqbsn]
	cqbsj := cqbsg + cqbsS[cqbsp].yys + 1

	if cqbsj >= cqbsLast {
		cqbsstate = cqbsAct[cqbsg]
	} else {
		cqbsstate = cqbsAct[cqbsj]
		if cqbsChk[cqbsstate] != -cqbsn {
			cqbsstate = cqbsAct[cqbsg]
		}
	}
	// dummy call; replaced with literal code
	switch cqbsnt {

	case 1:
		cqbsDollar = cqbsS[cqbspt-1 : cqbspt+1]
		//line query.y:107
		{
			cqbslex.(*Lexer).root = rootNode{Keys: keys, Tree: cqbsDollar[1].node}
		}
	case 2:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:113
		{
			cqbsVAL.node = andNode{Left: cqbsDollar[1].node, Right: cqbsDollar[3].node}
		}
	case 3:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:117
		{
			cqbsVAL.node = orNode{Left: cqbsDollar[1].node, Right: cqbsDollar[3].node}
		}
	case 4:
		cqbsDollar = cqbsS[cqbspt-2 : cqbspt+1]
		//line query.y:121
		{
			cqbsVAL.node = notNode{Clause: cqbsDollar[2].node}
		}
	case 5:
		cqbsDollar = cqbsS[cqbspt-1 : cqbspt+1]
		//line query.y:125
		{
			cqbsVAL.node = cqbsDollar[1].node
		}
	case 6:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:131
		{
			cqbsVAL.node = likeNode{Key: cqbsDollar[1].str, Value: cqbsDollar[3].str}
		}
	case 7:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:135
		{
			cqbsVAL.node = equalNode{Key: cqbsDollar[1].str, Value: cqbsDollar[3].str}
		}
	case 8:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:139
		{
			cqbsVAL.node = nequalNode{Key: cqbsDollar[1].str, Value: cqbsDollar[3].str}
		}
	case 9:
		cqbsDollar = cqbsS[cqbspt-2 : cqbspt+1]
		//line query.y:143
		{
			cqbsVAL.node = hasNode{Key: cqbsDollar[2].str}
		}
	case 10:
		cqbsDollar = cqbsS[cqbspt-3 : cqbspt+1]
		//line query.y:147
		{
			cqbsVAL.node = cqbsDollar[2].node
		}
	case 11:
		cqbsDollar = cqbsS[cqbspt-1 : cqbspt+1]
		//line query.y:153
		{
			keys = append(keys, cqbsDollar[1].str)
			cqbsVAL.str = cqbsDollar[1].str
		}
	case 12:
		cqbsDollar = cqbsS[cqbspt-1 : cqbspt+1]
		//line query.y:160
		{
			cqbsVAL.str = cqbsDollar[1].str[1 : len(cqbsDollar[1].str)-1]
		}
	}
	goto cqbsstack /* stack new state and value */
}
