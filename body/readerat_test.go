package body

import (
	"io"
	"testing"

	"github.com/chronos-tachyon/morehttp/internal/mockreader"
)

func TestReaderAtBody_WithSeekAndStat(t *testing.T) {
	fi := &mockreader.FileInfo{NameValue: "input.txt", SizeValue: 4, ModeValue: 0444}
	r := mockreader.New(
		mockreader.ExpectStat(fi, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectReadAt(nil, 0, 0, nil),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'b'}, 1, 1, nil),
		mockreader.ExpectReadAt([]byte{'c'}, 2, 1, nil),
		mockreader.ExpectReadAt([]byte{'d'}, 3, 1, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Seek-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("Seek-End"),
		mockreader.ExpectMark("ReadAt-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("ReadAt-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper111{Inner: r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}

func TestReaderAtBody_WithSeekNoStat(t *testing.T) {
	r := mockreader.New(
		mockreader.ExpectSeek(0, io.SeekEnd, 4, nil),
		mockreader.ExpectSeek(0, io.SeekStart, 0, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectReadAt(nil, 0, 0, nil),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'b'}, 1, 1, nil),
		mockreader.ExpectReadAt([]byte{'c'}, 2, 1, nil),
		mockreader.ExpectReadAt([]byte{'d'}, 3, 1, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Seek-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("Seek-End"),
		mockreader.ExpectMark("ReadAt-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("ReadAt-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper011{Inner: r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}

func TestReaderAtBody_WithStatNoSeek(t *testing.T) {
	fi := &mockreader.FileInfo{NameValue: "input.txt", SizeValue: 4, ModeValue: 0444}
	r := mockreader.New(
		mockreader.ExpectStat(fi, nil),
		mockreader.ExpectMark("ShortBody-Begin"),
		mockreader.ExpectMark("Read-Begin"),
		mockreader.ExpectReadAt(nil, 0, 0, nil),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'b'}, 1, 1, nil),
		mockreader.ExpectReadAt([]byte{'c'}, 2, 1, nil),
		mockreader.ExpectReadAt([]byte{'d'}, 3, 1, nil),
		mockreader.ExpectMark("Read-End"),
		mockreader.ExpectMark("Seek-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("Seek-End"),
		mockreader.ExpectMark("ReadAt-Begin"),
		mockreader.ExpectReadAt([]byte{'a'}, 0, 1, nil),
		mockreader.ExpectReadAt([]byte{'c', 'd'}, 2, 2, nil),
		mockreader.ExpectMark("ReadAt-End"),
		mockreader.ExpectMark("Close-Begin"),
		mockreader.ExpectClose(nil),
		mockreader.ExpectMark("Close-End"),
		mockreader.ExpectMark("AfterClose-Begin"),
		mockreader.ExpectMark("AfterClose-End"),
		mockreader.ExpectMark("ShortBody-End"),
	)

	b, err := FromReader(mockreader.Wrapper101{Inner: r})
	if err != nil {
		t.Errorf("FromReader failed: %v", err)
		return
	}

	RunBodyTests(t, &TestOptions{
		ShortMock: r,
		ShortBody: b,
	})
}
