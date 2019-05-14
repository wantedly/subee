package prefixer

import (
	"bytes"

	"golang.org/x/text/transform"
)

type PrefixFunc func() string

// New returns a new prefixer.
func New(prefixFunc PrefixFunc) transform.Transformer {
	return &transformerImpl{
		prefixFunc: prefixFunc,
	}
}

type transformerImpl struct {
	prefixed   bool
	prefixFunc PrefixFunc
}

// Reset implements transform.Transformer.Reset.
func (t *transformerImpl) Reset() {
	t.prefixed = false
}

// Transform implements transform.Transformer.Transform.
func (t *transformerImpl) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var buf bytes.Buffer
	var nWantDst int
	for _, chr := range src {
		if !t.prefixed {
			n, _ := buf.WriteString(t.prefixFunc() + " ")
			nWantDst += n
			t.prefixed = true
		}
		if chr == '\n' {
			t.prefixed = false
		}
		buf.WriteByte(chr)
		nWantDst++
		nSrc++
	}
	nDst = copy(dst, buf.Bytes())
	if nDst < nWantDst {
		err = transform.ErrShortDst
	}
	return
}
