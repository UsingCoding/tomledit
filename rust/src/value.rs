use crate::error::{invalid_raw_value, protocol};
use crate::protocol::{PathElement, ProtocolError, TomlValue};
use toml_edit::{Array, DocumentMut, Item, Value};

pub fn to_value(value: TomlValue, path: &[PathElement]) -> Result<Value, ProtocolError> {
    match value {
        TomlValue::String(value) => Ok(Value::from(value)),
        TomlValue::Integer(value) => Ok(Value::from(value)),
        TomlValue::Float(value) => {
            if !value.is_finite() {
                return Err(protocol("float value must be finite"));
            }
            Ok(Value::from(value))
        }
        TomlValue::Boolean(value) => Ok(Value::from(value)),
        TomlValue::Array(values) => {
            let mut array = Array::new();
            for value in values {
                array.push(to_value(value, path)?);
            }
            Ok(Value::Array(array))
        }
        TomlValue::Raw(raw) => parse_raw(raw, path),
    }
}

fn parse_raw(raw: String, path: &[PathElement]) -> Result<Value, ProtocolError> {
    let document: DocumentMut = format!("value = {raw}")
        .parse()
        .map_err(|error| invalid_raw_value(path, format!("parse raw TOML value: {error}")))?;
    let table = document.as_table();
    if table.len() != 1 {
        return Err(invalid_raw_value(
            path,
            "raw TOML must contain exactly one value",
        ));
    }
    match table.get("value") {
        Some(Item::Value(value)) => Ok(value.clone()),
        _ => Err(invalid_raw_value(path, "raw TOML is not a value")),
    }
}
