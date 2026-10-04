use tomledit_wasm::protocol::Request;

pub fn request(toml: &str, operations: serde_json::Value) -> Request {
    serde_json::from_value(serde_json::json!({
        "version": 1,
        "toml": toml,
        "operations": operations,
    }))
    .expect("valid test protocol request")
}
