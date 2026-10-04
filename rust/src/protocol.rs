use serde::{Deserialize, Serialize};

pub const VERSION: u8 = 1;

#[derive(Debug, Deserialize)]
pub struct Request {
    pub version: u8,
    pub toml: String,
    pub operations: Vec<Operation>,
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(tag = "type", content = "value", rename_all = "snake_case")]
pub enum PathElement {
    Key(String),
    Index(i64),
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(tag = "type", content = "value", rename_all = "snake_case")]
pub enum TomlValue {
    String(String),
    Integer(i64),
    Float(f64),
    Boolean(bool),
    Array(Vec<TomlValue>),
    Raw(String),
}

#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct Field {
    pub key: String,
    pub value: TomlValue,
}

#[derive(Debug, Deserialize)]
#[serde(tag = "op", rename_all = "snake_case")]
pub enum Operation {
    Set {
        path: Vec<PathElement>,
        value: TomlValue,
    },
    Delete {
        path: Vec<PathElement>,
    },
    ArrayAppend {
        path: Vec<PathElement>,
        value: TomlValue,
    },
    ArrayInsert {
        path: Vec<PathElement>,
        index: i64,
        value: TomlValue,
    },
    ArrayReplace {
        path: Vec<PathElement>,
        index: i64,
        value: TomlValue,
    },
    ArrayRemove {
        path: Vec<PathElement>,
        index: i64,
    },
    AotAppend {
        path: Vec<PathElement>,
        fields: Vec<Field>,
    },
    AotInsert {
        path: Vec<PathElement>,
        index: i64,
        fields: Vec<Field>,
    },
    AotReplace {
        path: Vec<PathElement>,
        index: i64,
        fields: Vec<Field>,
    },
    AotRemove {
        path: Vec<PathElement>,
        index: i64,
    },
}

#[derive(Serialize)]
pub struct Response {
    pub version: u8,
    pub ok: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub toml: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<ProtocolError>,
}

impl Response {
    pub fn success(toml: String) -> Self {
        Self {
            version: VERSION,
            ok: true,
            toml: Some(toml),
            error: None,
        }
    }

    pub fn failure(error: ProtocolError) -> Self {
        Self {
            version: VERSION,
            ok: false,
            toml: None,
            error: Some(error),
        }
    }
}

#[derive(Clone, Debug, Serialize)]
pub struct ProtocolError {
    pub code: String,
    pub message: String,
    pub path: Vec<PathElement>,
}
