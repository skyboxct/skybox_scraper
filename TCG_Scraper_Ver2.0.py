import sys
import gspread
import time
import requests
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials
from requests_html import HTMLSession

MAX_RETRIES = 2

da_output_range = ["A", "F"]
tcg_output_range = ["G", "K"]
tnt_output_range = ["L", "O"]
cs_output_range = ["P", "S"]


# "client_email": "tcgscraper@tcg-scraper.iam.gserviceaccount.com"
# TODO: Optimize Speed
# TODO: Work around JAVA for TCG Player

session = HTMLSession()

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
    da_row = tcg_row = tnt_row = cs_row = 2
    for start_index in range(0, len(urls), 40):
        batch_list = []
        for url in urls[start_index:start_index + 10]:  # The first item is a list containing all the information
            if url:
                r = session.get(url[0])
                r.html.render()
                soup = BeautifulSoup(r.html.raw_html, features="lxml")
                # soup = BeautifulSoup(response.content, features="lxml")

                if "www.dacardworld.com" in url[0]:
                    print("Reading url info for ", "line:", da_row, url[0])
                    output_range = da_output_range[0] + str(da_row) + ":" + da_output_range[1] + str(da_row)
                    values = retry(parse_da, soup, MAX_RETRIES)
                    da_row += 1
                elif "shop.tcgplayer.com" in url[0]:
                    print("Reading url info for ", "line:", tcg_row, url[0])
                    output_range = tcg_output_range[0] + str(tcg_row) + ":" + tcg_output_range[1] + str(tcg_row)
                    values = retry(parse_tcg, r, MAX_RETRIES)
                    tcg_row += 1
                    print(values)
                elif "www.trollandtoad.com" in url[0]:
                    print("Reading url info for ", "line:", tnt_row, url[0])
                    output_range = tnt_output_range[0] + str(tnt_row) + ":" + tnt_output_range[1] + str(tnt_row)
                    values = retry(parse_tnt, soup, MAX_RETRIES)
                    tnt_row += 1
                    print(values)
                elif "collectorstore.com" in url[0]:
                    print("Reading url info for ", "line:", cs_row, url[0])
                    output_range = cs_output_range[0] + str(cs_row) + ":" + cs_output_range[1] + str(cs_row)
                    values = retry(parse_cs, soup, MAX_RETRIES)
                    cs_row += 1
                    print(values)

                batch_list.append({'range': output_range, 'values': values})
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


def parse_tcg(r):
    tcg_title_field = r.html.find(".product-details__name", first=True)  # Grabs item Name
    if tcg_title_field is None:
        tcg_title = "Could not retrieve title"
    else:
        tcg_title = tcg_title_field.text

    tcg_price_field = r.html.find(".spotlight__price", first=True)
    if tcg_price_field is None:
        tcg_price = "Could not retrieve price"
    else:
        tcg_price = tcg_price_field.text

    tcg_desc_field = r.html.find(".pd-description__description", first=True) # Grabs Description
    if tcg_desc_field is None:
        tcg_desc = "Could not retrieve description"
    else:
        tcg_desc = tcg_desc_field.text

    tcg_pic_field =  r.html.find(".product-details__product", first=True)
    if tcg_pic_field is None:
        tcg_pic = "Could not retrieve pic"
    else:
        try:
            tcg_pic =tcg_pic_field.find(".progressive-image-main", first=True).attrs["src"]
        except KeyError:
            tcg_pic = "Could not retrieve pic"

    # logic for stock information
    tcg_stock_text = ""
    if tcg_price == "":
        tcg_stock_text = "Out Of Stock"
    else:
        tcg_stock_text = "In Stock"

    return [[tcg_title, tcg_price, tcg_desc, tcg_pic, tcg_stock_text]]


def parse_tnt(soup):
    tnt_title = soup.find('h1').get_text().strip()
    tnt_pics = soup.find(id='main-prod-img').findChild()['data-src']

    try:
        tnt_price = float(soup.find(class_="d-flex flex-column").span.get_text().strip('$'))
    except AttributeError:
        tnt_price = ""

    tnt_stock_text = ""

    # logic for stock information
    if tnt_price == "":
        tnt_stock_text = "Out of Stock"
    else:
        tnt_stock_text = "In Stock"

    return [[tnt_title, tnt_price, tnt_pics, tnt_stock_text]]


def parse_cs(soup):
    cs_title = soup.find('h1').get_text().strip()

    try:
        cs_price = float(soup.find(class_="price price--withoutTax").get_text().strip("$"))
    except AttributeError:
        cs_price = ""

    cs_pic = soup.find(class_='productView-img-container').findChild()['href']

    # logic for stock information
    if cs_price == "":
        cs_stock_text = "Out of Stock"
    else:
        cs_stock_text = "In Stock"

    return [[cs_title, cs_price, cs_pic, cs_stock_text]]


def retry(func, soup, max_tries):
    for i in range(0, max_tries):
        try:
            results = func(soup)
        except Exception as ex:
            template = "An exception of type {0} occurred. Arguments:\n{1!r}"
            message = template.format(type(ex).__name__, ex.args)
            print(message)
        else:
            return results

if __name__ == "__main__":
    main()
