import React, {Component} from 'react';
import useForm from "react-hook-form";
import {RHFInput} from 'react-hook-form-input';
import Socket from "./socket";
import Simulation from './Simulation'
// import Select from 'react-select';
import {Checkbox, MenuItem, Select} from '@material-ui/core';
import Input from "@material-ui/core/Input";

function LaneChoice(props) {
    var options = [
        {value: 0, label: 'uniform Lane choice'},
        {value: 1, label: 'trafficBasedChoice'},
    ];
    return <div> {props.name}: <RHFInput
        as={<Select>
            {options.map(option => <MenuItem value={option.value}>{option.label}</MenuItem>)}
        </Select>}
        rules={{required: true}}
        name={props.name}
        register={props.register}
        setValue={props.setValue}
    /></div>
}

function CarDistributionType(props) {
    var options = [
        {value: 0, label: 'exponentialDistribution'},
        {value: 1, label: 'normalDistribution'},
        {value: 2, label: 'poissonDistribution'},
        {value: 3, label: 'constantDistribution'},
        {value: 4, label: 'uniformDistribution'},
    ];
    return <div> {props.name}: <RHFInput
        as={<Select>
            {options.map(option => <MenuItem value={option.value}>{option.label}</MenuItem>)}
        </Select>}
        rules={{required: true}}
        name={props.name}
        register={props.register}
        setValue={props.setValue}
    /></div>
}

function SimulationDisplayForm(props) {
    const {register, handleSubmit, setValue} = useForm({
        defaultValues: {
            'displayCarDetails': true
        }
    });
    return <div>
        <br/>
        <h3> Simulation Display Details:</h3>
        displayCarDetails:
        <RHFInput
            onChange={handleSubmit(props.onSubmit)}
            as={<Checkbox/>}
            name="displayCarDetails"
            type="checkbox"
            register={register}
            setValue={setValue}
        />
        <br/>
    </div>
}

function SimulationForm(props) {
    const {register, handleSubmit, setValue} = useForm({
        defaultValues: {
            "sizeOfLane": 10,
            "numVerticalLanes": 1,
            "numHorizontalLanes": 1,
            'inAlpha': 1,
            'outBeta': 1,
            'carMovementP': 0.5,
            'probSwitchingLanes': 0.1,
            'accidentProb': 0,
            'numHorizontalCars': 10,
            'numVerticalCars': 10,
            'inLaneChoice': 0,
            'outLaneChoice': 0,
            'laneSwitchChoice': 0,
            'carRemovalRate': 1,
            'carRestartProb': 0.25,
            'carClock': 1,
            'carSpeedUniformEndRange': 0,
            'CarDistributionType': 0,
            'reSampleSpeedEveryClk': true,
            'probPolicePullOverProb': 0.01,
            'speedBasedPullOver': false,
            'parkingEnabled': false,
            'distractionRate': 0,
            'parkingTimeRate': 1,
            'crossWalkCutoff': 2,
            'crossWalkEnabled': false,
            'pedestrianDeathAccidentProb': 0.5,
            'probEnteringIntersection': 1,
            'intersectionAccidentProb': 0.1,
            'accidentScaling': true,
            'slowDownSpeed': 1,
            'removeUnlikelyEvents': true,
            'unlikelyCutoff': 0.05
        }
    }); // initialise the hook

    return (
        <div>
            <h3>Please Input Parameters for the simulation</h3>
            {/*{props.simulating && <button onClick={props.cancelSimulation}>Stop Simulation</button>}*/}
            <form onSubmit={handleSubmit(props.onSubmit)} className={"dataform"}>
                <input type="submit" disabled={props.simulating} value={"Submit"}/>
                <br/>
                Size Of Lane: <Input type={"text"} name="sizeOfLane" inputRef={register}/> {/* register an input */}
                <br/>
                Number of Horizontal Lanes: <Input type={"text"} name="numHorizontalLanes"
                                                   inputRef={register}/> {/* register an input */}
                <br/>
                Number of Vertical Lanes: <Input type={"text"} name="numVerticalLanes"
                                                 inputRef={register}/> {/* register an input */}
                <br/>
                inAlpha: <Input type={"text"} name="inAlpha"
                                inputRef={register} step="any"/> {/* register an input */}
                <br/>
                outBeta: <Input type={"text"} name="outBeta"
                                inputRef={register} step="any"/> {/* register an input */}
                <br/>
                carMovementP: <Input type={"text"} name="carMovementP"
                                     inputRef={register} step="any"/> {/* register an input */}
                <br/>
                numHorizontalCars: <Input type={"text"} name="numHorizontalCars"
                                          inputRef={register} step="any"/> {/* register an input */}
                <br/>
                numVerticalCars: <Input type={"text"} name="numVerticalCars"
                                        inputRef={register} step="any"/> {/* register an input */}
                <br/>
                <LaneChoice name={"inLaneChoice"} register={register} setValue={setValue}/>
                <br/>
                <LaneChoice name={"outLaneChoice"} register={register} setValue={setValue}/>
                <br/>

                probSwitchingLanes: <Input type={"text"} name="probSwitchingLanes"
                                           inputRef={register} step="any"/> {/* register an input */}
                <br/>
                <LaneChoice name={"laneSwitchChoice"} register={register} setValue={setValue}/>

                <br/>
                accidentProb: <Input type={"text"} name="accidentProb"
                                     inputRef={register} step="any"/> {/* register an input */}
                <br/>

                carRemovalRate: <Input type={"text"} name="carRemovalRate"
                                       inputRef={register} step="any"/> {/* register an input */}
                <br/>
                carRestartProb: <Input type={"text"} name="carRestartProb"
                                       inputRef={register} step="any"/> {/* register an input */}
                <br/>

                carClock: <Input type={"text"} name="carClock"
                                 inputRef={register} step="any"/> {/* register an input */}
                <br/>

                carSpeedUniformEndRange: <Input type={"text"} name="carSpeedUniformEndRange"
                                                inputRef={register} step="any"/> {/* register an input */}
                <br/>

                <CarDistributionType name={"CarDistributionType"} register={register} setValue={setValue}/>
                <br/>

                reSampleSpeedEveryClk:
                <RHFInput
                    as={<Checkbox/>}
                    name="reSampleSpeedEveryClk"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>

                probPolicePullOverProb: <Input type={"text"} name="probPolicePullOverProb"
                                               inputRef={register} step="any"/> {/* register an input */}
                <br/>
                speedBasedPullOver:
                <RHFInput
                    as={<Checkbox/>}
                    name="speedBasedPullOver"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>

                parkingEnabled:
                <RHFInput
                    as={<Checkbox/>}
                    name="parkingEnabled"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>
                distractionRate: <Input type={"text"} name="distractionRate"
                                        inputRef={register} step="any"/> {/* register an input */}
                <br/>
                parkingTimeRate: <Input type={"text"} name="parkingTimeRate"
                                        inputRef={register} step="any"/> {/* register an input */}
                <br/>
                crossWalkCutoff: <Input type={"text"} name="crossWalkCutoff"
                                        inputRef={register} step="any"/> {/* register an input */}
                <br/>

                crossWalkEnabled:
                <RHFInput
                    as={<Checkbox/>}
                    name="crossWalkEnabled"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>

                pedestrianDeathAccidentProb: <Input type={"text"} name="pedestrianDeathAccidentProb"
                                                    inputRef={register} step="any"/> {/* register an input */}
                <br/>

                probEnteringIntersection: <Input type={"text"} name="probEnteringIntersection"
                                                 inputRef={register} step="any"/> {/* register an input */}
                <br/>
                intersectionAccidentProb: <Input type={"text"} name="intersectionAccidentProb"
                                                 inputRef={register} step="any"/> {/* register an input */}
                <br/>
                accidentScaling:
                <RHFInput
                    as={<Checkbox/>}
                    name="accidentScaling"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>
                slowDownSpeed: <Input type={"text"} name="slowDownSpeed"
                                      inputRef={register} step="any"/> {/* register an input */}
                <br/>

                remove Unlikely events:
                <RHFInput
                    as={<Checkbox/>}
                    name="removeUnlikelyEvents"
                    type="checkbox"
                    register={register}
                    setValue={setValue}
                />
                <br/>

                Unlikely Cutoff:
                <Input type={"text"} name="unlikelyCutoff"
                       inputRef={register} step="any"/> {/* register an input */}
                <br/>


                {/*<input type="submit" disabled={props.simulating} value={"Submit"}/>*/}
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
            displayCarDetails: true
        }
    }

    // componentDidMount is a react life-cycle method that runs after the component
    //   has mounted.
    componentDidMount() {
        // establish websocket connection to backend server.
        var environment = process.env.NODE_ENV === 'production' ? 'production' : 'development';
        // Had to manually add this because https://github.com/parcel-bundler/parcel/issues/496#issuecomment-366993459
        let ws = new WebSocket('ws://localhost:5000/ws');

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
        // console.log(data.locations)
    };

    completedSimulation = () => {
        console.log("Simulation completed event");
        this.setState({simulating: false})
    };

    cancelSimulation = () => {
        this.setState({simulating: false});
        this.socket.emit('cancelSimulation', 'cancel simulation');
    };

    startSimulation = (event) => {
        console.log('Starting SimulationHandler with params');
        event["id"] = this.state.clientId;
        this.socket.emit('startSimulation', {
            "event": "startSimulation",
            "data": event
        });
    };

    changeDisplayDetails = (event) => {
        console.log("set car details to", event.displayCarDetails)
        this.setState({
            displayCarDetails: event.displayCarDetails
        })
    }

    render() {
        return (
            <div className="App">
                {this.state.simulating &&
                <Simulation simulating={this.state.simulating} data={this.state.simulationData}
                            displayCarDetails={this.state.displayCarDetails}/>}
                <div>Running Simulation: {this.state.simulating.toString()}</div>
                <SimulationForm onSubmit={this.startSimulation} simulating={this.state.simulating}
                                cancelSimulation={this.cancelSimulation}/>
                <SimulationDisplayForm onSubmit={this.changeDisplayDetails}/>
            </div>
        );
    }
}

export default SimulationHandler;