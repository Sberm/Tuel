// Tuel
// Toolset (English, libc, CUDA, OpenGL)
// Catergorization (?)
// Format of a tool (tool (tool name), descr (description))
// Build: Search, Implement
//   Search - Full text search on tool and descr
//   Implement - Natural language, pseudo code, highlighted keywords
mod db;
mod site;

fn main() {
    db::conn::conn();
    site::serve();
}
