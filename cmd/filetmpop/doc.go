// Copyright 2016 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The command filetmpop starts a server.
//
// Usage:
//
//
//	$ filetmpop -h
// 		Usage of dist/darwin-amd64/filetmpop:
// 		  -addr string
// 			Listen address (default "tcp://0.0.0.0:46658")
// 		  -path string
// 			path to directory where files are stored (default "/var/stratumn/filestore")
// 		  -tmsp string
// 			TMSP server: socket | grpc (default "socket")
//
// Docker:
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 46658:46658 stratumn/filetmpop

package main
