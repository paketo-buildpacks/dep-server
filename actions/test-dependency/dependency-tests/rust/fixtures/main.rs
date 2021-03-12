use std::env;
use std::io::prelude::*;
use std::net::{TcpListener, TcpStream};

fn main() {
    let args: Vec<String> = env::args().collect();
    let port = &args[1];

    let listener = TcpListener::bind(format!("127.0.0.1:{}", port)).unwrap();

    for stream in listener.incoming() {
        let stream = stream.unwrap();
        println!("Connection established!");

        handle_connection(stream);
    }
}

fn handle_connection(mut stream: TcpStream) {
    let mut buffer = [0; 1024];

    stream.read(&mut buffer).unwrap();

    let response = "Hello, world!";

    stream.write(response.as_bytes()).unwrap();
    stream.flush().unwrap();
}
