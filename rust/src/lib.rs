pub mod apply;
pub mod error;
pub mod path;
pub mod protocol;
pub mod value;

use apply::apply_request;
use protocol::{ProtocolError, Request, Response};

#[unsafe(no_mangle)]
pub extern "C" fn alloc(len: u32) -> u32 {
    let mut bytes = vec![0_u8; len as usize];
    debug_assert_eq!(bytes.len(), bytes.capacity());
    let pointer = bytes.as_mut_ptr();
    std::mem::forget(bytes);
    pointer as u32
}

/// Frees a buffer previously returned by [`alloc`] or [`apply`].
///
/// # Safety
///
/// `pointer` and `len` must be the exact pair returned by this module, and
/// the buffer must not have been freed already.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn dealloc(pointer: u32, len: u32) {
    if pointer == 0 && len == 0 {
        return;
    }
    // SAFETY: alloc and apply return buffers with capacity equal to len, and the
    // Go ABI returns each buffer exactly once with its original length.
    unsafe {
        drop(Vec::from_raw_parts(
            pointer as *mut u8,
            len as usize,
            len as usize,
        ));
    }
}

/// Applies one JSON request stored in linear memory and returns a packed buffer.
///
/// # Safety
///
/// `pointer` and `len` must identify a live buffer returned by [`alloc`] for
/// the duration of this call.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn apply(pointer: u32, len: u32) -> u64 {
    // SAFETY: Go provides a buffer allocated by alloc for the duration of this call.
    let input = unsafe { std::slice::from_raw_parts(pointer as *const u8, len as usize) };
    let response = match serde_json::from_slice::<Request>(input) {
        Ok(request) => match apply_request(request) {
            Ok(toml) => Response::success(toml),
            Err(error) => Response::failure(error),
        },
        Err(error) => Response::failure(ProtocolError {
            code: "protocol_error".into(),
            message: format!("decode request: {error}"),
            path: Vec::new(),
        }),
    };
    let bytes = serde_json::to_vec(&response).unwrap_or_else(|_| {
        br#"{\"version\":1,\"ok\":false,\"error\":{\"code\":\"protocol_error\",\"message\":\"serialize response\",\"path\":[]}}"#.to_vec()
    });
    response_buffer(bytes)
}

fn response_buffer(bytes: Vec<u8>) -> u64 {
    let mut bytes = bytes.into_boxed_slice().into_vec();
    debug_assert_eq!(bytes.len(), bytes.capacity());
    let len = bytes.len() as u32;
    let pointer = bytes.as_mut_ptr() as u32;
    std::mem::forget(bytes);
    (u64::from(pointer) << 32) | u64::from(len)
}
