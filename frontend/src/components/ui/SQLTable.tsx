import type { Component } from 'solid-js'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './table'

type SQLTableProps = {
  columns: string[]
  rows: string[][]
  queryType?: string
  executionTime?: string
  rowsCount?: number
}

const SQLTable: Component<SQLTableProps> = props => {
  return (
    <div class="w-full h-full flex flex-col">
      <div class="flex-shrink-0 p-2 border-b bg-muted/10 text-xs text-muted-foreground">
        <div class="flex items-center gap-4">
          <span>Query Type: {props.queryType || 'SELECT'}</span>
          {props.executionTime && <span>Execution Time: {props.executionTime}</span>}
          {props.rowsCount !== undefined && <span>Rows: {props.rowsCount}</span>}
        </div>
      </div>

      {/* Table Container with Sticky Header */}
      <div class="flex-1 overflow-auto relative">
        <Table>
          <TableHeader class="sticky top-0 z-10 bg-background border-b">
            <TableRow>
              {props.columns.map(column => (
                <TableHead class="font-mono text-md font-bold bg-background text-foreground">
                  {column}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.rows.map(row => (
              <TableRow class="group">
                {row.map(cell => (
                  <TableCell class="font-mono text-xs text-foreground/70 group-hover:text-foreground/90 transition-colors duration-200">
                    {cell}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

export default SQLTable
