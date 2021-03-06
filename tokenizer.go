package dfm

import "unicode"

func newTokenizer(code []rune) tokenizer {
	return tokenizer{
		code: code,
		line: 1,
		col:  1,
	}
}

type tokenizer struct {
	code []rune
	cur  int
	line int
	col  int
}

func (t *tokenizer) next() token {
	haveType := tokenIllegal
	start := t.cur
	line, col := t.line, t.col

	digit := func(r rune) bool {
		return '0' <= r && r <= '9'
	}

	r := t.currentRune()
	switch r {
	case 0:
		return token{
			tokenType: tokenEOF,
			line:      line,
			col:       col,
		}
	case '+', '-', '[', ']', '(', ')', '{', '}', '<', '>', '=', ':', '.', ',':
		t.nextRune()
		haveType = tokenType(r)
	case '\'':
		for {
			r := t.nextRune()
			if r == 0 {
				break
			}
			if r == '\'' {
				if t.nextRune() != '\'' {
					break
				}
			}
		}
		haveType = tokenString
	case '#':
		for digit(t.nextRune()) {
		}
		haveType = tokenCharacter
	default:
		if unicode.IsSpace(r) {
			for unicode.IsSpace(t.nextRune()) {
			}
			haveType = tokenWhiteSpace
		} else if r == '_' || unicode.IsLetter(r) {
			word := func(r rune) bool {
				return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
			}
			for word(t.nextRune()) {
			}
			haveType = tokenWord
		} else if digit(r) {
			haveType = tokenInteger
			for digit(t.nextRune()) {
			}
			if t.currentRune() == '.' {
				t.nextRune()
				haveType = tokenFloat
				for digit(t.nextRune()) {
				}
			}
			if t.currentRune() == 'e' || t.currentRune() == 'E' {
				t.nextRune()
				haveType = tokenFloat
				if t.currentRune() == '+' || t.currentRune() == '-' {
					t.nextRune()
				}
				for digit(t.nextRune()) {
				}
			}
		} else {
			// For an illegal token we only consume one rune. The next call to
			// this function then tries to continue after the illegal rune.
			t.nextRune()
		}
	}

	return token{
		tokenType: haveType,
		text:      string(t.code[start:t.cur]),
		line:      line,
		col:       col,
	}
}

// findClosingBrace is an optimization for reading binary data. They are written
// as hex characters like this:
//
//     Bitmap.Data = {
//       ABCDEF0123456789
//       ABCDEF0123456789
//       ABCDEF0123456789}
//
// which would be tokenized to integers and words. Since binary data is the
// largest part of a typical DFM, this function allows the parser to process
// this much quicker than re-combining integers and words.
func (t *tokenizer) findClosingBrace() []rune {
	oldLine, oldCol := t.line, t.col

	for i := t.cur; i < len(t.code); i++ {
		if t.code[i] == '\n' {
			t.line++
			t.col = 1
		} else {
			t.col++
		}

		if t.code[i] == '}' {
			part := t.code[t.cur:i]
			t.cur = i
			return part
		}
	}

	// No closing brace was found.
	t.line, t.col = oldLine, oldCol
	return nil
}

func (t *tokenizer) currentRune() rune {
	if t.cur < len(t.code) {
		return t.code[t.cur]
	}
	return 0
}

func (t *tokenizer) nextRune() rune {
	if t.cur < len(t.code) {
		if t.code[t.cur] == '\n' {
			t.line++
			t.col = 1
		} else {
			t.col++
		}
		t.cur++
	}
	return t.currentRune()
}
