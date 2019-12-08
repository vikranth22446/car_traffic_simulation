import React, {Component} from 'react';

class Car extends Component {
    render() {
        return (
            <div>
                🚗
                <div className={"car title"}>{this.props.details.name}</div>
                {this.props.displayCarDetails && <div className={"cardetails"}>
                    <div className={"speed"}>Speed: {this.props.details.speed}</div>
                    <div className={"waitingTime"}>Waiting Time: {this.props.details.waitingTime}</div>
                </div>}
            </div>
        );
    }
}

class SimulationCars extends Component {
    render() {
        var items = [];
        var value;
        // console.log(this.props.car);
        for (var key in this.props.car) {
            value = this.props.car[key];
            items.push({"name": key, "waitingTime": value.WaitingTime, "speed": value.Speed})
        }

        if (this.props.locationState === 0) {
            return <td className={"empty"}>
                <div/>
            </td>
        }

        if (this.props.locationState === 3) {
            return <td className={"parking"}>
                <td>
                    <div>
                        {items.map(item => <div><Car details={item} displayCarDetails={this.props.displayCarDetails}/>
                        </div>)}
                    </div>
                </td>
            </td>
        }

        if (this.props.locationState === 4) {
            return <td>
                <div>
                    🔥
                    🔥
                    {items.map(item => <div>🔥<Car details={item} displayCarDetails={this.props.displayCarDetails}/>🔥
                    </div>)}
                    🔥
                    🔥
                </div>
            </td>
        }

        return (
            <td>
                <div>

                    {items.map(item => <div><Car details={item} displayCarDetails={this.props.displayCarDetails}/>
                    </div>)}
                </div>
            </td>
        );


        return <td>
            <div>&nbsp;&nbsp;&nbsp;&nbsp;</div>
        </td>
    }
}

class Simulation extends Component {
    render() {
        return <table className={"tableBorder"}>
            {this.props.data.map(row => {
                return <tr>
                    {row.map(item => {
                        return <SimulationCars locationState={item.state} car={item.cars}
                                               displayCarDetails={this.props.displayCarDetails}/>
                    })}
                </tr>
            })}
        </table>
    }
}

export default Simulation
