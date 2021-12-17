package body

import (
	"io"
	"testing"

	"github.com/chronos-tachyon/morehttp/internal/mockreader"
)

func TestBufferedBody_WithStatNoReadAtNoSeek(t *testing.T) {
	fi := &mockreader.FileInfo{NameValue: "input.txt", SizeValue: 4, ModeValue: 0444}
	p := make([]byte, 4)
	p[0] = 'a'
	p[1] = 'b'
	p[2] = 'c'
	p[3] = 'd'
	r := mockreader.New(
		mockreader.ExpectStat(fi, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectRead(p, 4, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper100{r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}

func TestBufferedBody_WithReadAtNoSeekNoStat(t *testing.T) {
	p := make([]byte, 65536)
	p[0] = 'a'
	p[1] = 'b'
	p[2] = 'c'
	p[3] = 'd'
	r := mockreader.New(
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectRead(p, 4, io.EOF),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper001{r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock:              r,
		ShortBody:              b,
		ShortBodyUnknownLength: true,
	})
}
