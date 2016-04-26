import sphinx
import strutils
import docopt

const doc = """
Generate fixture data for the Sphinx library - output data created by
sphinx_add_query as lines of 4, 2-digit hexadecimal integers.

Usage:
  fixture_data <query> [<index>] [<comment>]
  fixture_data --dump-response <query> [<index>] [<comment>]
  fixture_data --dump-header <query> [<index>] [<comment>]
  fixture_data (-h | --help)
  fixture_data (-v | --version)

Options:
  -h --help     Show this screen.
  -v --version  Show version.
"""

type whatToDump = enum
  header,
  response,
  request

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

proc get_header_buffer(client: PClient) : cstring {.cdecl, importc:"sphinx_dump_header"
dynlib: sphinxDll.}

proc get_header_length(client: PClient) : cint {.cdecl, importc:"sphinx_header_length"
dynlib: sphinxDll.}

proc main(query, index, comment: string, dumpInfo: whatToDump) =
  let client = sphinx.create(copy_args = false)

  discard client.add_query(query, index, comment)

  case dumpInfo:
  of request:
    for i in 1 .. client.get_num_requests():
      var (buffer, buflen) = ( client.get_nth_request(i), client.get_nth_request_length(i) )

      if buffer.isNil: continue

      # Header of buffer length
      echo buflen

      for j in 0 .. <buflen:
        if j == 0:
          printf("%02x:",cast[cuchar](buffer[j]))
          continue
        # Every 4 characters need to print newline
        case j mod 4
        of 0:
          printf("\n%02x:",cast[cuchar](buffer[j]))
        of 3:
          printf("%02x",cast[cuchar](buffer[j]))
        else:
          printf("%02x:",cast[cuchar](buffer[j]))
  of response:
    # TODO: Implementation of dumping response buffer
    echo "Dump response"
  of header:
    var (buffer,bufLen) = (client.get_header_buffer(),client.get_header_length())
    if buffer.isNil:
      echo "Error getting header buffer from Sphinx API"
      client.destroy()
      quit(1)

    echo bufLen

    for i in 0 .. <bufLen:
      if i == 0: printf("%02x:",cast[cuchar](buffer[i])); continue

      case i mod 4
      of 0: printf("\n%02x:",cast[cuchar](buffer[i]))
      of 3: printf("%02x",cast[cuchar](buffer[i]))
      else: printf("%02x:",cast[cuchar](buffer[i]))

  client.destroy()

when isMainModule:
  let
    args = docopt(doc, version = "1.0")
    query   = if args["<query>"]: $args["<query>"] else: ""
    index   = if args["<index>"]: $args["<index>"] else: "*"
    comment = if args["<comment>"]: $args["<comment>"] else: ""

  let outputDump: whatToDump =
    if args["--dump-response"]:
      whatToDump.response
    elif args["--dump-header"]:
      whatToDump.header
    else:
      whatToDump.request

  #echo "Making query with `$#` `$#` `$#`" % [query, index, comment]
  main( query, index, comment, outputDump )
