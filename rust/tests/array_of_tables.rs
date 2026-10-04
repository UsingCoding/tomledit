mod common;

use common::request;
use tomledit_wasm::apply::apply_request;

const ACTIONS: &str = "[[foo.items]]\nname = \"one\"\ngroup = \"one\"\n\n[[foo.items]]\nname = \"two\"\ngroup = \"two\"\n";

fn fields(name: &str) -> serde_json::Value {
    serde_json::json!([
        {"key":"name","value":{"type":"string","value":name}},
        {"key":"group","value":{"type":"string","value":"one"}}
    ])
}

#[test]
fn applies_all_aot_mutations_and_nested_index_path() {
    let output = apply_request(request(
        ACTIONS,
        serde_json::json!([
            {"op":"aot_append","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"fields":fields("four")},
            {"op":"aot_insert","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"index":1,"fields":fields("inserted")},
            {"op":"aot_replace","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"index":2,"fields":fields("replaced")},
            {"op":"aot_remove","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"index":3},
            {"op":"set","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"},{"type":"index","value":1},{"type":"key","value":"group"}],"value":{"type":"string","value":"three"}}
        ]),
    ))
    .expect("array-of-tables operations succeed");
    assert!(output.contains("name = \"inserted\"\ngroup = \"three\""));
    assert!(output.contains("name = \"replaced\""));
    assert!(!output.contains("name = \"four\""));

    let duplicate = apply_request(request(
        ACTIONS,
        serde_json::json!([{"op":"aot_append","path":[{"type":"key","value":"foo"},{"type":"key","value":"items"}],"fields":[{"key":"name","value":{"type":"string","value":"a"}},{"key":"name","value":{"type":"string","value":"b"}}]}]),
    ))
    .expect_err("duplicate fields fail");
    assert_eq!(duplicate.code, "protocol_error");
}
