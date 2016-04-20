import sphinx
import strutils
import docopt

const doc = """
Generate fixture data for the Sphinx library - output data created by
sphinx_add_query as lines of 4, 2-digit hexadecimal integers.

Usage:
  fixture_data <query> [<index>] [<comment>]
  fixture_data (-h | --help)
  fixture_data (-v | --version)

Options:
  -h --help     Show this screen.
  -v --version  Show version.
"""

converter toSPHBool(b: bool): SphinxBool = result = if b: SPH_TRUE else: SPH_FALSE

proc printf(formatstr: cstring) {.importc: "printf", varargs,
                                    header: "<stdio.h>".}

# Need to provide own wrappers for debugging functions that were added to the
# Sphinx library
proc get_num_requests(client: PClient): cint {.cdecl, importc: "sphinx_get_num_requests",
dynlib: sphinxDll.}

proc get_nth_request(client: PClient, n: cint): cstring {.cdecl, importc: "sphinx_get_nth_request",
dynlib: sphinxDll.}

proc get_nth_request_length(client: PClient, n: cint): cint {.cdecl, importc: "sphinx_get_nth_request_length"
dynlib: sphinxDll.}

proc main(query, index, comment: string) =
  let client = sphinx.create(copy_args = false)

  discard client.add_query(query, index, comment)

  for i in 1 .. client.get_num_requests():
    var (buffer, buflen) = ( client.get_nth_request(i), client.get_nth_request_length(i) )

    if buffer.isNil: continue

    var
      finished = false
      pos = 0

    while not finished:
      # line number
      # printf("%d\t",(pos div 4) + 1)
      #
      # FIXME: Fix this to make sure that we're outputting all bytes from buffer - not
      # guaranteed to be a multiple of 4.
      printf(
        "%02x:%02x:%02x:%02x\n",
        cast[cuchar](buffer[pos]),
        cast[cuchar](buffer[pos + 1]),
        cast[cuchar](buffer[pos + 2]),
        cast[cuchar](buffer[pos + 3]),
      )

      if pos+3 >= buflen:
        finished = true
        break

      pos += 4

  client.destroy()

when isMainModule:
  let
    args = docopt(doc, version = "1.0")
    query   = if args["<query>"]: $args["<query>"] else: ""
    index   = if args["<index>"]: $args["<index>"] else: "*"
    comment = if args["<comment>"]: $args["<comment>"] else: ""

  #echo "Making query with `$#` `$#` `$#`" % [query, index, comment]
  main( query, index, comment )
