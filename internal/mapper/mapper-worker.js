/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

// This code will run in the web worker

// Listen for messages from the main thread
self.addEventListener("message", (event) => {
    const currentUrl = event.data.currentUrl;
    const urls = event.data.urls;

    // Return if there is no data left to send
    if (urls.length === 0) {
        return;
    }

    // Send the URLs to Cartograph Mapper
    const data = {source: currentUrl, destinations: urls};
    fetch(currentUrl, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Cartograph": "mapper-data",
        },
        body: JSON.stringify(data),
    })
        .then((response) =>
            self.postMessage(
                `Sent URLs to Cartograph Mapper with response: ${response.status}`
            )
        )
        .catch((error) =>
            console.error(`Error sending URLs to Cartograph Mapper: ${error}`)
        );
});
