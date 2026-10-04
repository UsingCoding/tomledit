use crate::protocol::{PathElement, ProtocolError};

pub fn invalid_toml(message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "invalid_toml".into(),
        message: message.into(),
        path: Vec::new(),
    }
}

pub fn path_not_found(path: &[PathElement], message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "path_not_found".into(),
        message: message.into(),
        path: path.to_vec(),
    }
}

pub fn type_mismatch(path: &[PathElement], message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "type_mismatch".into(),
        message: message.into(),
        path: path.to_vec(),
    }
}

pub fn index_out_of_range(path: &[PathElement], message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "index_out_of_range".into(),
        message: message.into(),
        path: path.to_vec(),
    }
}

pub fn invalid_raw_value(path: &[PathElement], message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "invalid_raw_value".into(),
        message: message.into(),
        path: path.to_vec(),
    }
}

pub fn protocol(message: impl Into<String>) -> ProtocolError {
    ProtocolError {
        code: "protocol_error".into(),
        message: message.into(),
        path: Vec::new(),
    }
}
