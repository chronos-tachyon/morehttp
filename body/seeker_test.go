package body

import (
	"io"
	"testing"

	"github.com/chronos-tachyon/morehttp/internal/mockreader"
)

func TestSeekerBody_WithStat(t *testing.T) {
	fi := &mockreader.FileInfo{NameValue: "input.txt", SizeValue: 4, ModeValue: 0444}
	r := mockreader.New(
		mockreader.ExpectStat(fi, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead(nil, 0, nil),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(1, io.SeekStart, 1, nil),
		mockreader.ExpectRead([]byte{'b'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c'}, 1, nil),
		mockreader.ExpectSeek(3, io.SeekStart, 3, nil),
		mockreader.ExpectRead([]byte{'d'}, 1, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Seek-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c', 'd'}, 2, nil),
		mockreader.ExpectMark("Seek-End"),
		mockreader.ExpectMark("ReadAt-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c', 'd'}, 2, nil),
		mockreader.ExpectMark("ReadAt-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper110{r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}

func TestSeekerBody_NoStat(t *testing.T) {
	r := mockreader.New(
		mockreader.ExpectSeek(0, io.SeekEnd, 4, nil),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead(nil, 0, nil),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(1, io.SeekStart, 1, nil),
		mockreader.ExpectRead([]byte{'b'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c'}, 1, nil),
		mockreader.ExpectSeek(3, io.SeekStart, 3, nil),
		mockreader.ExpectRead([]byte{'d'}, 1, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Seek-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c', 'd'}, 2, nil),
		mockreader.ExpectMark("Seek-End"),
		mockreader.ExpectMark("ReadAt-Begin"),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectRead([]byte{'a'}, 1, nil),
		mockreader.ExpectSeek(2, io.SeekStart, 2, nil),
		mockreader.ExpectRead([]byte{'c', 'd'}, 2, nil),
		mockreader.ExpectMark("ReadAt-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper010{r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}
