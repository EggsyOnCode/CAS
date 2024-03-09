# Distributed CAS (Content Addressed Storage)

Implemented in Golang

We have a network of nodes each running this server on their systems; they can now save a file not only on their local disks but also on other nodes

Similarly they can retrieve their file from other nodes usign the key against which they had saved the file in the network.

All files are stored on remote peer/nodes in an encrypted fashion so they can't view their peer's private data.

