import sphinx

converter toSPHBool(b: bool): SphinxBool = result = if b: SPH_TRUE else: SPH_FALSE

converter toBool(sph: SphinxBool): bool =
  case sph:
  of SPH_TRUE: true
  of SPH_FALSE: false

let client = sphinx.create(copy_args = false)

echo "Created client successfully"

client.destroy()


