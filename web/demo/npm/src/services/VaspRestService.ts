import {VaspDetails} from "../models/VaspDetails";

export interface VaspsResponse {
    data: VaspDetails[];
}

export class VaspRestService {
    restServiceUrl: string

    constructor(restServiceUrl: string) {
        this.restServiceUrl = restServiceUrl
    }

    //function
    getVasps(): Promise<VaspDetails[]> {
        return fetch(this.restServiceUrl + '/vasps')
            .then(res => res.json())
            .then(res => res as VaspsResponse)
            .then(res => res.data)
    }
}
