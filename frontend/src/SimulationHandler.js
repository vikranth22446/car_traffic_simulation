import React, {Component} from 'react';
import useForm from "react-hook-form";
import Socket from "./socket";
import Simulation from './Simulation'

function SimulationForm(props) {
    const {register, handleSubmit} = useForm({
        defaultValues: {
            "sizeOfLane": 10,
            "numVerticalLanes": 1,
            "numHorizontalLanes": 1,
        }
    }); // initialise the hook

    return (
        <div>
            <form onSubmit={handleSubmit(props.onSubmit)}>
                Size Of Lane: <input type={"number"} name="sizeOfLane" ref={register}/> {/* register an input */}
                <br/>
                Number of Horizontal Lanes: <input type={"number"} name="numHorizontalLanes"
                                                   ref={register}/> {/* register an input */}
                <br/>
                Number of Vertical Lanes: <input type={"number"} name="numVerticalLanes"
                                                 ref={register}/> {/* register an input */}
                <br/>
                <input type="submit"/>
            </form>
        </div>
    )
}

class SimulationHandler extends Component {
    // TODO handle connection
    // TODO handle getting lanes
    // TODO handle rendering lanes in smth of a table format

    // Note: We could handle this as canvas, but it seems to easier to just render tables
    constructor(props) {
        super(props);

        // state variables
        this.state = {
            connected: false,
            sizeOfLane: props.sizeOfLane,
            simulating: false,
            simulationData: new Array([]),
            clientId: null,
        }
    }

    // componentDidMount is a react life-cycle method that runs after the component
    //   has mounted.
    componentDidMount() {
        // establish websocket connection to backend server.
        let ws = new WebSocket('ws://localhost:80/ws');

        // create and assign a socket to a variable.
        let socket = this.socket = new Socket(ws);

        // handle connect and discconnect events.
        socket.on('connect', this.onConnect);
        socket.on('disconnect', this.onDisconnect);

        /* EVENT LISTENERS */
        // event listener to handle 'hello' from a server
        socket.on('identify', this.receiveIdentification); // Example event here would be
        socket.on('simulationUpdate', this.updateSimulationState); // Example event here would be
        socket.on('completedSimulation', this.completedSimulation); // Example event here would be
    }

    // onConnect sets the state to true indicating the socket has connected
    //    successfully.
    onConnect = () => {
        this.setState({connected: true});
    };

    // onDisconnect sets the state to false indicating the socket has been
    //    disconnected.
    onDisconnect = () => {
        this.setState({connected: false});
    };

    receiveIdentification = (data) => {
        console.log("Identification recieved", data);
        this.setState({clientId: data})
    };
    // helloFromClient is an event emitter that sends a hello message to the backend
    //    server on the socket.
    helloFromClient = () => {
        console.log('saying hello...');
        this.socket.emit('helloFromClient', 'hello server!');
    };

    // updateSimulationState is an event listener/consumer that handles hello messages
    //    from the backend server on the socket.
    updateSimulationState = (data) => {
        console.log("update simulation event");
        this.setState({simulating: true, simulationData: data.locations});
    };

    completedSimulation = () => {
        console.log("Simulation completed event");
        this.setState({simulating: false})
    };

    startSimulation = (event) => {
        console.log('Starting SimulationHandler with params');
        this.socket.emit('startSimulation', {
            "event": "startSimulation",
            "data": {
                "id": this.state.clientId,
                "sizeOfLane": event.sizeOfLane,
                "numHorizontalLanes": event.numHorizontalLanes,
                "numVerticalLanes": event.numVerticalLanes
            }
        });

    };


    render() {
        return (
            <div className="App">
                <Simulation simulating={this.state.simulating} data={this.state.simulationData}/>
                <div>Please Input Parameters for the simulation</div>
                <SimulationForm onSubmit={this.startSimulation}/>
            </div>
        );
    }
}

export default SimulationHandler;