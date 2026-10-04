use std::collections::HashSet;

use crate::error::{invalid_toml, protocol, type_mismatch};
use crate::path::{checked_index, checked_insert_index, table_for_terminal};
use crate::protocol::{Field, Operation, PathElement, ProtocolError, Request, VERSION};
use crate::value::to_value;
use toml_edit::{DocumentMut, Item, Table};

pub fn apply_request(request: Request) -> Result<String, ProtocolError> {
    if request.version != VERSION {
        return Err(protocol(format!(
            "unsupported request version {}",
            request.version
        )));
    }
    let mut document: DocumentMut = request
        .toml
        .parse()
        .map_err(|error| invalid_toml(format!("parse TOML: {error}")))?;
    for operation in request.operations {
        apply_operation(document.as_table_mut(), operation)?;
    }
    Ok(document.to_string())
}

fn apply_operation(root: &mut Table, operation: Operation) -> Result<(), ProtocolError> {
    match operation {
        Operation::Set { path, value } => {
            let value = to_value(value, &path)?;
            let (table, key) = table_for_terminal(root, &path)?;
            if let Some(Item::Value(existing)) = table.get_mut(&key) {
                let decor = existing.decor().clone();
                *existing = value;
                *existing.decor_mut() = decor;
            } else {
                table.insert(&key, Item::Value(value));
            }
            Ok(())
        }
        Operation::Delete { path } => {
            let (table, key) = table_for_terminal(root, &path)?;
            if table.remove(&key).is_none() {
                return Err(crate::error::path_not_found(
                    &path,
                    format!("key {key:?} does not exist"),
                ));
            }
            Ok(())
        }
        Operation::ArrayAppend { path, value } => {
            let value = to_value(value, &path)?;
            let array = array_at(root, &path)?;
            array.push(value);
            Ok(())
        }
        Operation::ArrayInsert { path, index, value } => {
            let value = to_value(value, &path)?;
            let array = array_at(root, &path)?;
            let index = checked_insert_index(index, array.len(), &path)?;
            array.insert(index, value);
            Ok(())
        }
        Operation::ArrayReplace { path, index, value } => {
            let value = to_value(value, &path)?;
            let array = array_at(root, &path)?;
            let index = checked_index(index, array.len(), &path)?;
            array.replace(index, value);
            Ok(())
        }
        Operation::ArrayRemove { path, index } => {
            let array = array_at(root, &path)?;
            let index = checked_index(index, array.len(), &path)?;
            array.remove(index);
            Ok(())
        }
        Operation::AotAppend { path, fields } => {
            let table = make_table(fields, &path)?;
            let tables = aot_at(root, &path)?;
            tables.push(table);
            Ok(())
        }
        Operation::AotInsert {
            path,
            index,
            fields,
        } => {
            let table = make_table(fields, &path)?;
            let tables = aot_at(root, &path)?;
            let index = checked_insert_index(index, tables.len(), &path)?;
            tables.insert(index, table);
            Ok(())
        }
        Operation::AotReplace {
            path,
            index,
            fields,
        } => {
            let table = make_table(fields, &path)?;
            let tables = aot_at(root, &path)?;
            let index = checked_index(index, tables.len(), &path)?;
            tables.replace(index, table);
            Ok(())
        }
        Operation::AotRemove { path, index } => {
            let tables = aot_at(root, &path)?;
            let index = checked_index(index, tables.len(), &path)?;
            tables.remove(index);
            Ok(())
        }
    }
}

fn array_at<'a>(
    root: &'a mut Table,
    path: &[PathElement],
) -> Result<&'a mut toml_edit::Array, ProtocolError> {
    let (table, key) = table_for_terminal(root, path)?;
    let item = table
        .get_mut(&key)
        .ok_or_else(|| crate::error::path_not_found(path, format!("key {key:?} does not exist")))?;
    item.as_value_mut()
        .and_then(toml_edit::Value::as_array_mut)
        .ok_or_else(|| type_mismatch(path, format!("key {key:?} is not an array")))
}

fn aot_at<'a>(
    root: &'a mut Table,
    path: &[PathElement],
) -> Result<&'a mut toml_edit::ArrayOfTables, ProtocolError> {
    let (table, key) = table_for_terminal(root, path)?;
    let item = table
        .get_mut(&key)
        .ok_or_else(|| crate::error::path_not_found(path, format!("key {key:?} does not exist")))?;
    item.as_array_of_tables_mut()
        .ok_or_else(|| type_mismatch(path, format!("key {key:?} is not an array of tables")))
}

fn make_table(fields: Vec<Field>, path: &[PathElement]) -> Result<Table, ProtocolError> {
    let mut seen = HashSet::with_capacity(fields.len());
    let mut table = Table::new();
    for field in fields {
        if !seen.insert(field.key.clone()) {
            return Err(protocol(format!(
                "duplicate array-of-tables field key {:?}",
                field.key
            )));
        }
        table.insert(&field.key, Item::Value(to_value(field.value, path)?));
    }
    Ok(table)
}
