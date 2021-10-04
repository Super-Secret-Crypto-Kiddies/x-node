import argparse
import subprocess

parser = argparse.ArgumentParser(description="Host an XChain node.")

parser.add_argument('-c', '--chains', help="Which blockchain nodes to run (e.g. btc eth doge)", type=str, nargs='+', dest='chains')

args = parser.parse_args()

# node_compose_profiles = ['btc', 'eth', 'doge', 'bch', 'xmr', 'zec'] # init values are the defaults
profiles = []
defaults = ['btc', 'eth']

nodes = {
    'btc': 8333,
    'eth': 30303,
    'bch': 8333,
    'doge': 22556,
    'ltc': 9333,
    'xmr': 18080,
    'zec': 8233
}

profiles_args = []
if args.chains != None:
    if 'btc' in args.chains and 'bch' in args.chains:
        raise Exception("BTC and BCH nodes both use conflicting port 8333. Please choose only one.")
    for c in args.chains:
        if nodes[c] == None:
            profiles_args = ""
            raise Exception(f"Invalid blockchain node type '{c}'")
        else:
            profiles_args += ["--profile", c]
    profiles = args.chains
else:
    for c in defaults:
        if c != 'bch':
            profiles_args += ["--profile", c]

print(f"Running on ports: {' '.join([str(nodes[node]) for node in profiles])}")
subprocess.run(["docker", "compose", *profiles_args, "up"])