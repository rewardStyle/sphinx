import sphinx
import strutils

converter toSPHBool(b: bool): SphinxBool = result = if b: SPH_TRUE else: SPH_FALSE

converter toBool(sph: SphinxBool): bool =
  case sph:
  of SPH_TRUE: true
  of SPH_FALSE: false

# Need to provide own wrappers for debugging functions that were added to the
# Sphinx library
proc get_num_requests(client: PClient): cint {.cdecl, importc: "sphinx_get_num_requests",
dynlib: sphinxDll.}

proc get_nth_request(client: PClient, n: cint): pointer {.cdecl, importc: "sphinx_get_nth_request",
dynlib: sphinxDll.}

proc get_nth_request_length(client: PClient, n: cint): cint {.cdecl, importc: "sphinx_get_nth_request_length"
dynlib: sphinxDll.}

let client = sphinx.create(copy_args = false)

echo "Created client successfully"
echo "Starting number of requests: $#" % $(client.get_num_requests())

var
  query_index: int = client.add_query(query="What up?",index_list="*",comment="comment1")

if query_index != -1:
  echo "Added query successfully!"

echo "Number of requests now: $#" % $(client.get_num_requests)

echo "Request buffer length: $#" % $(client.get_nth_request_length(1))

client.destroy()
