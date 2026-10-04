use crate::error::{index_out_of_range, path_not_found, type_mismatch};
use crate::protocol::{PathElement, ProtocolError};
use toml_edit::{Item, Table};

pub fn table_for_terminal<'a>(
    root: &'a mut Table,
    path: &[PathElement],
) -> Result<(&'a mut Table, String), ProtocolError> {
    let Some(PathElement::Key(terminal)) = path.last() else {
        return Err(type_mismatch(path, "terminal path element must be a key"));
    };
    let table = descend_table(root, &path[..path.len() - 1], path)?;
    Ok((table, terminal.clone()))
}

fn descend_table<'a>(
    table: &'a mut Table,
    elements: &[PathElement],
    full_path: &[PathElement],
) -> Result<&'a mut Table, ProtocolError> {
    let Some((element, remaining)) = elements.split_first() else {
        return Ok(table);
    };
    let PathElement::Key(key) = element else {
        return Err(type_mismatch(
            full_path,
            "array index cannot select a table",
        ));
    };
    let item = table
        .get_mut(key)
        .ok_or_else(|| path_not_found(full_path, format!("key {key:?} does not exist")))?;
    match item {
        Item::Table(next) => descend_table(next, remaining, full_path),
        Item::ArrayOfTables(tables) => {
            let Some((PathElement::Index(index), after_index)) = remaining.split_first() else {
                return Err(type_mismatch(
                    full_path,
                    "array of tables requires an index path element",
                ));
            };
            let index = checked_index(*index, tables.len(), full_path)?;
            let next = tables.get_mut(index).ok_or_else(|| {
                index_out_of_range(
                    full_path,
                    format!("index {index} is outside array of tables"),
                )
            })?;
            descend_table(next, after_index, full_path)
        }
        _ => Err(type_mismatch(
            full_path,
            format!("key {key:?} is not a table or array of tables"),
        )),
    }
}

pub fn checked_index(
    index: i64,
    length: usize,
    path: &[PathElement],
) -> Result<usize, ProtocolError> {
    let index = usize::try_from(index).map_err(|_| {
        index_out_of_range(path, "index must be non-negative and within the collection")
    })?;
    if index >= length {
        return Err(index_out_of_range(
            path,
            format!("index {index} is outside collection of length {length}"),
        ));
    }
    Ok(index)
}

pub fn checked_insert_index(
    index: i64,
    length: usize,
    path: &[PathElement],
) -> Result<usize, ProtocolError> {
    let index = usize::try_from(index).map_err(|_| {
        index_out_of_range(path, "index must be non-negative and within the collection")
    })?;
    if index > length {
        return Err(index_out_of_range(
            path,
            format!("index {index} is outside insertion range 0..={length}"),
        ));
    }
    Ok(index)
}
