import gspread
import pyppeteer
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

session = HTMLSession()


def main():
    print("Starting Scraper")

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
    scrape_url_list(total_list, sheet)


def scrape_url_list(urls, sheet):
    # Dictionary to hold host configurations and row 'cursor'
    host_configs = {
        "DA": {"range": da_output_range, "row": 2, "parser_func": parse_da, "parser_param": None},
        "TCG": {"range": tcg_output_range, "row": 2, "parser_func": parse_tcg, "parser_param": None},
        "TNT": {"range": tnt_output_range, "row": 2, "parser_func": parse_tnt, "parser_param": None},
        "CS": {"range": cs_output_range, "row": 2, "parser_func": parse_cs, "parser_param": None}
    }

    for start_index in range(0, len(urls), 10):
        batch_list = []
        for url in urls[start_index:start_index + 10]:  # The first item is a list containing all the information
            if url:
                try:
                    host = get_host(url[0])
                    print("Reading url info for ", "line:", host_configs[host]["row"], url[0])

                    response_html = retry(session.get, url[0], MAX_RETRIES)
                    response_html.html.render()
                except pyppeteer.errors.TimeoutError:
                    # TODO: Fix frequent timeouts to TNT
                    print("Error: request timed out to", url[0])
                    host_configs[host]["row"] += 1
                    continue
                except KeyError:
                    print("configuration not found for host: ", host, "\nURL:", url[0])
                    continue

                if host == "TCG":
                    host_configs[host]["parser_param"] = response_html
                else:
                    host_configs[host]["parser_param"] = BeautifulSoup(response_html.html.raw_html, features="lxml")

                output_range = host_configs[host]["range"][0] + str(host_configs[host]["row"]) + ":" + host_configs[host]["range"][1] + str(host_configs[host]["row"])
                values = retry(host_configs[host]["parser_func"], host_configs[host]["parser_param"], MAX_RETRIES)
                host_configs[host]["row"] += 1
                print(values)

                if output_range and values:
                    batch_list.append({'range': output_range, 'values': values})

        print("Pushing lines", start_index, "-", start_index + 10, "to sheet")
        sheet.batch_update(batch_list)
    # TODO: roll up parsing errors into output before exit


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


# TODO: fix missing field issues
def parse_tcg(r):
    tcg_title_field = r.html.find(".product-details__name", first=True)  # Grabs item Name
    if tcg_title_field is None:
        print("Error: title not found in html")
        tcg_title = ""
    else:
        tcg_title = tcg_title_field.text

    tcg_price_field = r.html.find(".spotlight__price", first=True)  # Grabs price
    if tcg_price_field is None:
        print("Error: price not found in html")
        tcg_price = ""
    else:
        tcg_price = tcg_price_field.text

    tcg_desc_field = r.html.find(".pd-description__description", first=True)  # Grabs Description
    if tcg_desc_field is None:
        print("Error: description not found in html")
        tcg_desc = ""
    else:
        tcg_desc = tcg_desc_field.text

    tcg_pic_field = r.html.find(".product-details__product", first=True)  # Grabs main picture
    if tcg_pic_field is None:
        print("Error: pic field not found in html")
        tcg_pic = ""
    else:
        try:
            tcg_pic = tcg_pic_field.find(".progressive-image-main", first=True).attrs["src"]
        except KeyError:
            print("Error: source link not found in pic field")
            tcg_pic = ""

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


def retry(func, param, max_retries):
    for i in range(0, max_retries+1):
        try:
            results = func(param)
        except Exception as ex:
            template = "An exception of type {0} occurred. Arguments:\n{1!r}"
            message = template.format(type(ex).__name__, ex.args)
        else:
            return results
    print("Parse failed after", i+1, "attempts:", message)


def get_host(url):
    if "dacardworld.com" in url:
        return "DA"
    elif "tcgplayer.com" in url:
        return "TCG"
    elif "trollandtoad.com" in url:
        return "TNT"
    elif "collectorstore.com" in url:
        return "CS"
    else:
        print("Unknown host for url:", url)
        return "Unknown"

if __name__ == "__main__":
    main()
