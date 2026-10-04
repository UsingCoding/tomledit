package tomledit_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/usingcoding/tomledit"
)

func TestSet(t *testing.T) {
	output := apply(t, "[package]\nversion = \"1\"\n", tomledit.Set(tomledit.P(tomledit.K("package"), tomledit.K("version")), tomledit.String("2")))
	if output != "[package]\nversion = \"2\"\n" {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestDelete(t *testing.T) {
	output := apply(t, "name = \"demo\"\ndeprecated = true\n", tomledit.Delete(tomledit.P(tomledit.K("deprecated"))))
	if output != "name = \"demo\"\n" {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestArrayAppend(t *testing.T) {
	output := apply(t, fixture(t, "arrays.toml"), tomledit.ArrayAppend(tomledit.P(tomledit.K("tools")), tomledit.String("git")))
	if output != "tools = [\"go\", \"cargo\", \"git\"]\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestArrayInsert(t *testing.T) {
	output := apply(t, fixture(t, "arrays.toml"), tomledit.ArrayInsert(tomledit.P(tomledit.K("tools")), 1, tomledit.String("lint")))
	if output != "tools = [\"go\", \"lint\", \"cargo\"]\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestArrayReplace(t *testing.T) {
	output := apply(t, fixture(t, "arrays.toml"), tomledit.ArrayReplace(tomledit.P(tomledit.K("tools")), 1, tomledit.String("rust")))
	if output != "tools = [\"go\", \"rust\"]\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestArrayRemove(t *testing.T) {
	output := apply(t, fixture(t, "arrays.toml"), tomledit.ArrayRemove(tomledit.P(tomledit.K("tools")), 1))
	if output != "tools = [\"go\"]\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestAOTAppend(t *testing.T) {
	output := apply(t, fixture(t, "aot.toml"), tomledit.AOTAppend(itemsPath(), itemFields("new")...))
	assertOrdered(t, output, "name = \"qux\"", "name = \"new\"")
}

func TestAOTInsert(t *testing.T) {
	output := apply(t, fixture(t, "aot.toml"), tomledit.AOTInsert(itemsPath(), 2,
		tomledit.F("name", tomledit.String("new")),
		tomledit.F("value", tomledit.String("new")),
		tomledit.F("group", tomledit.String("foo")),
	))
	assertOrdered(t, output, "name = \"bar\"", "name = \"new\"", "name = \"baz\"")
}

func TestAOTReplace(t *testing.T) {
	output := apply(t, fixture(t, "aot.toml"), tomledit.AOTReplace(itemsPath(), 1, itemFields("new")...))
	if !strings.Contains(output, "name = \"new\"") || strings.Contains(output, "name = \"bar\"") {
		t.Fatalf("array-of-tables entry was not replaced:\n%s", output)
	}
}

func TestAOTRemove(t *testing.T) {
	output := apply(t, fixture(t, "aot.toml"), tomledit.AOTRemove(itemsPath(), 1))
	if strings.Contains(output, "name = \"bar\"") {
		t.Fatalf("array-of-tables entry was not removed:\n%s", output)
	}
}

func TestRawValue(t *testing.T) {
	output := apply(t, "channel = \"stable\"\n", tomledit.Set(tomledit.P(tomledit.K("channel")), tomledit.Raw("'nightly'")))
	if output != "channel = 'nightly'\n" {
		t.Fatalf("unexpected output: %q", output)
	}

	editor := newEditor(t)
	_, err := editor.Apply(context.Background(), []byte("channel = \"stable\"\n"), tomledit.Set(tomledit.P(tomledit.K("channel")), tomledit.Raw("1\nother = 2")))
	assertCode(t, err, tomledit.ErrInvalidRawValue)
}

func TestBatch(t *testing.T) {
	input := "version = \"1\"\nplugins = [\"git\"]\n\n[[foo.items]]\nname = \"bar\"\nvalue = \"bar\"\ngroup = \"foo\"\n"
	output := apply(t, input,
		tomledit.Set(tomledit.P(tomledit.K("version")), tomledit.String("2")),
		tomledit.ArrayAppend(tomledit.P(tomledit.K("plugins")), tomledit.String("golang")),
		tomledit.AOTInsert(itemsPath(), 0, itemFields("foo")...),
	)
	assertOrdered(t, output, "version = \"2\"", "name = \"foo\"", "name = \"bar\"")

	editor := newEditor(t)
	outputBytes, err := editor.Apply(context.Background(), []byte(input),
		tomledit.Set(tomledit.P(tomledit.K("version")), tomledit.String("2")),
		tomledit.ArrayRemove(tomledit.P(tomledit.K("plugins")), 4),
	)
	if outputBytes != nil {
		t.Fatalf("failed batch returned partial output: %q", outputBytes)
	}
	assertCode(t, err, tomledit.ErrIndexOutOfRange)
}

func TestInvalidTOML(t *testing.T) {
	editor := newEditor(t)
	_, err := editor.Apply(context.Background(), []byte("broken = [\n"))
	assertCode(t, err, tomledit.ErrInvalidTOML)
}

func TestPathNotFound(t *testing.T) {
	editor := newEditor(t)
	_, err := editor.Apply(context.Background(), []byte("name = \"demo\"\n"), tomledit.Set(tomledit.P(tomledit.K("missing"), tomledit.K("name")), tomledit.String("x")))
	assertCode(t, err, tomledit.ErrPathNotFound)
}

func TestTypeMismatch(t *testing.T) {
	editor := newEditor(t)
	_, err := editor.Apply(context.Background(), []byte("name = \"demo\"\n"), tomledit.ArrayAppend(tomledit.P(tomledit.K("name")), tomledit.String("x")))
	assertCode(t, err, tomledit.ErrTypeMismatch)
}

func TestIndexOutOfRange(t *testing.T) {
	editor := newEditor(t)
	_, err := editor.Apply(context.Background(), []byte("tools = [\"go\"]\n"), tomledit.ArrayRemove(tomledit.P(tomledit.K("tools")), 1))
	assertCode(t, err, tomledit.ErrIndexOutOfRange)
}

func TestFormattingPreserved(t *testing.T) {
	output := apply(t, fixture(t, "formatting.toml"), tomledit.AOTInsert(itemsPath(), 1, itemFields("middle")...))
	for _, fragment := range []string{
		"name    =  'demo'   # important",
		"[foo] # foo config\nenabled=true",
		"# keep me\n[[foo.items]]\nname = 'two'",
		"[unrelated]\nvalue={foo=\"bar\"}",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("unrelated formatting fragment changed: %q\noutput:\n%s", fragment, output)
		}
	}
}

func TestConcurrentApply(t *testing.T) {
	editor := newEditor(t)
	const workers = 16
	var group sync.WaitGroup
	errCh := make(chan error, workers)
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			output, err := editor.Apply(context.Background(), []byte("version = \"1\"\n"), tomledit.Set(tomledit.P(tomledit.K("version")), tomledit.String("2")))
			if err != nil {
				errCh <- err
				return
			}
			if string(output) != "version = \"2\"\n" {
				errCh <- errors.New("unexpected concurrent output")
			}
		}()
	}
	group.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

func newEditor(t *testing.T) *tomledit.Editor {
	t.Helper()
	editor, err := tomledit.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := editor.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	return editor
}

func apply(t *testing.T, input string, operations ...tomledit.Operation) string {
	t.Helper()
	output, err := newEditor(t).Apply(context.Background(), []byte(input), operations...)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func itemsPath() tomledit.Path {
	return tomledit.P(tomledit.K("foo"), tomledit.K("items"))
}

func itemFields(name string) []tomledit.Field {
	return []tomledit.Field{
		tomledit.F("name", tomledit.String(name)),
		tomledit.F("value", tomledit.String(name)),
		tomledit.F("group", tomledit.String("foo")),
	}
}

func assertOrdered(t *testing.T, output string, fragments ...string) {
	t.Helper()
	position := 0
	for _, fragment := range fragments {
		next := strings.Index(output[position:], fragment)
		if next < 0 {
			t.Fatalf("missing fragment %q in output:\n%s", fragment, output)
		}
		position += next + len(fragment)
	}
}

func assertCode(t *testing.T, err error, want tomledit.ErrorCode) {
	t.Helper()
	var editErr *tomledit.Error
	if !errors.As(err, &editErr) {
		t.Fatalf("expected *tomledit.Error, got %T: %v", err, err)
	}
	if editErr.Code != want {
		t.Fatalf("error code = %q, want %q (%v)", editErr.Code, want, err)
	}
}
