// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

import { ErrorUnauthorized } from '@/api/errors/ErrorUnauthorized';
import { HttpClient } from '@/utils/httpClient';
import { ABHitAction, ABTestApi, ABTestValues } from '@/types/abtesting';

/**
 * ABHttpApi is a console AB testing API.
 * Exposes all ab-testing related functionality
 */
export class ABHttpApi implements ABTestApi {
    private readonly http: HttpClient = new HttpClient();
    private readonly ROOT_PATH: string = '/api/v0/ab';

    /**
     * Used to get test banner information.
     *
     * @throws Error
     */
    public async fetchABTestValues(): Promise<ABTestValues> {
        const path = `${this.ROOT_PATH}/values`;
        const response = await this.http.get(path);
        if (response.ok) {
            const abResponse = await response.json();

            return new ABTestValues(
                abResponse.has_new_banner,
            );
        }

        if (response.status === 401) {
            throw new ErrorUnauthorized();
        }

        // use default values on error
        return new ABTestValues();
    }

    public async sendHit(action: ABHitAction): Promise<void> {
        const path = `${this.ROOT_PATH}/hit/${action}`;
        await this.http.post(path, null);
    }
}
