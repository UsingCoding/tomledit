mod common;

use common::request;
use tomledit_wasm::apply::apply_request;

#[test]
fn preserves_unrelated_comments_spacing_and_quote_style() {
    let input = "# project configuration\n\nname    =  'demo'   # important\n\n[foo] # foo config\nenabled=true\n\n\n[[foo.items]]\nname   = \"one\"\nvalue = \"one:run\"\ngroup  = \"foo\"\n\n# keep me\n[[foo.items]]\nname = 'two'\nvalue = \"two:run\"\ngroup = \"bar\"\n\n[unrelated]\nvalue={foo=\"bar\"}\n";
    let output = apply_request(request(
        input,
        serde_json::json!([{"op":"aot_insert","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"index":1,"fields":[{"key":"name","value":{"type":"string","value":"middle"}},{"key":"value","value":{"type":"string","value":"middle"}},{"key":"group","value":{"type":"string","value":"foo"}}]}]),
    ))
    .expect("formatted document is editable");

    assert!(output.contains("name    =  'demo'   # important"));
    assert!(output.contains("[foo] # foo config\nenabled=true"));
    assert!(output.contains("# keep me\n[[foo.items]]\nname = 'two'"));
    assert!(output.contains("[unrelated]\nvalue={foo=\"bar\"}"));
}
