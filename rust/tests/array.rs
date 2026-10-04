mod common;

use common::request;
use tomledit_wasm::apply::apply_request;

#[test]
fn applies_all_array_mutations() {
    let output = apply_request(request(
        "tools = [\"go\", \"cargo\"]\n",
        serde_json::json!([
            {"op":"array_append","path":[{"type":"key","value":"tools"}],"value":{"type":"string","value":"git"}},
            {"op":"array_insert","path":[{"type":"key","value":"tools"}],"index":1,"value":{"type":"string","value":"lint"}},
            {"op":"array_replace","path":[{"type":"key","value":"tools"}],"index":2,"value":{"type":"string","value":"rust"}},
            {"op":"array_remove","path":[{"type":"key","value":"tools"}],"index":3}
        ]),
    ))
    .expect("array operations succeed");
    assert_eq!(output, "tools = [\"go\", \"lint\", \"rust\"]\n");
}

#[test]
fn rejects_invalid_array_indices() {
    let error = apply_request(request(
        "tools = [\"go\"]\n",
        serde_json::json!([{"op":"array_replace","path":[{"type":"key","value":"tools"}],"index":1,"value":{"type":"string","value":"rust"}}]),
    ))
    .expect_err("out of range array replacement fails");
    assert_eq!(error.code, "index_out_of_range");
}
