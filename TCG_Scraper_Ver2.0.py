import sys
import gspread
import time
import requests
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials

da_output_range = ["A", "F"]
tcg_output_range = ["G", "K"]
tnt_output_range = ["L", "O"]
cs_output_range = ["P", "S"]


# "client_email": "tcgscraper@tcg-scraper.iam.gserviceaccount.com"
# TODO: Optimize Speed
# TODO: Work around JAVA for TCG Player

def main():
    print("Starting Scraper")

    headers = {'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:76.0) Gecko/20100101 Firefox/76.0'}
    scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
    creds = ServiceAccountCredentials.from_json_keyfile_name('tcgcreds.json', scope)
    client = gspread.authorize(creds)
    spreadsheet = client.open('Square TCG Scraper Ver1.1')
    sheet = spreadsheet.get_worksheet(1)

    rows = sheet.get_all_records()
    # Get all of the URLs in one object
    da_url_list = sheet.batch_get(['Y2:Y' + str(len(rows) + 1)])[0]
    tcg_url_list = sheet.batch_get(['Z2:Z' + str(len(rows) + 1)])[0]
    tnt_url_list = sheet.batch_get(['AA2:AA' + str(len(rows) + 1)])[0]
    cs_url_list = sheet.batch_get(['AB2:AB' + str(len(rows) + 1)])[0]

    total_list = da_url_list + tcg_url_list + tnt_url_list + cs_url_list
    scrape_url_list(total_list, sheet, headers)


def scrape_url_list(urls, sheet, headers):
    row = 2
    for start_index in range(0, len(urls), 40):
        batch_list = []
        for url in urls[start_index:start_index + 10]:  # The first item is a list containing all the information

            # TODO: check URL host, create host-specific parse funcs
            if url:
                print("Reading url info for ", "line:", row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                if "www.dacardworld.com" in url[0]:
                    output_range = da_output_range[0] + str(row) + ":" + da_output_range[1] + str(row)
                    values = parse_da(soup)
                    batch_list.append({'range': output_range, 'values': values})
                # TODO: reset row to 2
                elif "shop.tcgplayer.com" in url[0]:
                    output_range = tcg_output_range[0] + str(row) + ":" + tcg_output_range[1] + str(row)
                    values = parse_tcg(soup)
                    print(values)
                elif "www.trollandtoad.com" in url[0]:
                    output_range = tnt_output_range[0] + str(row) + ":" + tnt_output_range[1] + str(row)
                    values = parse_tnt(soup)
                    print(values)
                elif "collectorstore.com" in url[0]:
                    output_range = cs_output_range[0] + str(row) + ":" + cs_output_range[1] + str(row)
                    values = parse_tnt(soup)
                    print(values)

            row += 1
        sheet.batch_update(batch_list)


def parse_da(soup):
    title = soup.find('h1').get_text().strip()
    try:
        # Grab price info
        da_price = float(soup.find(class_='price large').get_text().strip("$").replace(',', ''))
    except AttributeError:
        da_price = ""

    try:
        # Need to get around no eight columns 'NoneType' Object
        da_desc = soup.find(class_='eight columns').get_text().strip()  # grab description
    except AttributeError:
        da_desc = soup.find(id='moredetailsTab').get_text().strip()

    da_pic = soup.find(class_='panel radius').findChild()['href']  # grab pic info
    da_upc = "UPC/Barcode" in soup.find(id='itemdetailsTab').find_all("li")[3].get_text()
    da_upc_text = ""
    da_stock_text = ""

    # logic for barcode information
    if da_upc:
        da_upc_text = soup.find(id='itemdetailsTab').find_all("li")[3].get_text().strip("UPC/Barcode:")

    # logic for stock information
    if da_price == "":
        da_stock_text = "Out Of Stock"
    else:
        da_stock_text = "In Stock"

    return [[title, da_price, da_desc, da_pic, da_upc_text, da_stock_text]]


def parse_tcg(soup):
    return [1, 2, 3, 4]


def parse_tnt(soup):
    return [5, 6, 7, 8]


def parse_cs(soup):
    return [9, 10, 11, 12]


if __name__ == "__main__":
    main()
