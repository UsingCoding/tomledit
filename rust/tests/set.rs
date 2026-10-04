mod common;

use common::request;
use tomledit_wasm::apply::apply_request;

#[test]
fn replaces_scalar_and_deletes_key() {
    let output = apply_request(request(
        "[package]\nname = \"demo\"\nversion = '1.0.0' # retain\n",
        serde_json::json!([
            {"op":"set","path":[{"type":"key","value":"package"},{"type":"key","value":"version"}],"value":{"type":"string","value":"2.0.0"}},
            {"op":"delete","path":[{"type":"key","value":"package"},{"type":"key","value":"name"}]}
        ]),
    ))
    .expect("operations succeed");

    assert!(output.contains("version = \"2.0.0\" # retain"));
    assert!(!output.contains("name ="));
}

#[test]
fn validates_raw_values_and_input_toml() {
    let raw = apply_request(request(
        "channel = \"stable\"\n",
        serde_json::json!([{"op":"set","path":[{"type":"key","value":"channel"}],"value":{"type":"raw","value":"'nightly'"}}]),
    ))
    .expect("valid raw value");
    assert_eq!(raw, "channel = 'nightly'\n");

    let invalid_raw = apply_request(request(
        "channel = \"stable\"\n",
        serde_json::json!([{"op":"set","path":[{"type":"key","value":"channel"}],"value":{"type":"raw","value":"1\nother = 2"}}]),
    ))
    .expect_err("multiple synthetic items are invalid raw");
    assert_eq!(invalid_raw.code, "invalid_raw_value");

    let invalid_toml =
        apply_request(request("bad = [\n", serde_json::json!([]))).expect_err("invalid TOML fails");
    assert_eq!(invalid_toml.code, "invalid_toml");
}

#[test]
fn rejects_missing_and_wrong_paths() {
    let missing = apply_request(request(
        "name = \"demo\"\n",
        serde_json::json!([{"op":"set","path":[{"type":"key","value":"missing"},{"type":"key","value":"name"}],"value":{"type":"string","value":"x"}}]),
    ))
    .expect_err("missing intermediate key fails");
    assert_eq!(missing.code, "path_not_found");

    let wrong_type = apply_request(request(
        "name = \"demo\"\n",
        serde_json::json!([{"op":"set","path":[{"type":"key","value":"name"},{"type":"key","value":"value"}],"value":{"type":"string","value":"x"}}]),
    ))
    .expect_err("scalar cannot be traversed");
    assert_eq!(wrong_type.code, "type_mismatch");
}
