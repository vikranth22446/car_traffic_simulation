import React, {Component} from "react";
import "./App.css";


class App extends Component {
    constructor(props) {
        super(props);

        // state variables
        this.state = {
            connected: false,
        }
    }

    // componentDidMount is a react life-cycle method that runs after the component
    //   has mounted.
    componentDidMount() {
        // establish websocket connection to backend server.
        let ws = new WebSocket('ws://localhost:4000');

        // create and assign a socket to a variable.
        let socket = this.socket = new Socket(ws);

        // handle connect and discconnect events.
        socket.on('connect', this.onConnect);
        socket.on('disconnect', this.onDisconnect);

        /* EVENT LISTENERS */
        // event listener to handle 'hello' from a server
        socket.on('helloFromServer', this.helloFromServer); // Example event here would be
    }

    // onConnect sets the state to true indicating the socket has connected
    //    successfully.
    onConnect = () => {
        this.setState({connected: true});
    }

    // onDisconnect sets the state to false indicating the socket has been
    //    disconnected.
    onDisconnect = () => {
        this.setState({connected: false});
    }

    // helloFromClient is an event emitter that sends a hello message to the backend
    //    server on the socket.
    helloFromClient = () => {
        console.log('saying hello...');
        this.socket.emit('helloFromClient', 'hello server!');
    }

    // helloFromServer is an event listener/consumer that handles hello messages
    //    from the backend server on the socket.
    helloFromServer = (data) => {
        console.log('hello from server! message:', data);
    };


    render() {
        const input = '# This is a header\n\nAnd this is a paragraph'
        return (
            <div className="App">
                <div className="Card">
                    <Upload/>
                </div>
                <div className="Section">
                    <Gallery/>
                </div>
            </div>
        );
    }
}

export default App;
