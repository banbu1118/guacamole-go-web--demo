import Guacamole from 'guacamole-common-js'

const mouse = function (element) {

    let guac_mouse = this;

    this.touchMouseThreshold = 3;
    this.scrollThreshold = 53;
    this.PIXELS_PER_LINE = 18;
    this.PIXELS_PER_PAGE = this.PIXELS_PER_LINE * 16;

    this.currentState = new Guacamole.Mouse.State(
        0, 0,
        false, false, false, false, false
    );

    this.onmousedown = null;
    this.onmouseup = null;
    this.onmousemove = null;
    this.onmouseout = null;

    var ignore_mouse = 0;
    var scroll_delta = 0;

    function cancelEvent(e) {
        e.stopPropagation();
        if (e.preventDefault) e.preventDefault();
        e.returnValue = false;
    }

    element.addEventListener("contextmenu", function (e) {
        cancelEvent(e);
    }, false);

    element.addEventListener("mousemove", function (e) {
        if (ignore_mouse) {
            ignore_mouse--;
            return;
        }
        guac_mouse.currentState.fromClientPosition(element, e.clientX, e.clientY);
        if (guac_mouse.onmousemove)
            guac_mouse.onmousemove(guac_mouse.currentState);
    }, false);

    element.addEventListener("mousedown", function (e) {
        cancelEvent(e);
        if (ignore_mouse)
            return;
        switch (e.button) {
            case 0:
                guac_mouse.currentState.left = true;
                break;
            case 1:
                guac_mouse.currentState.middle = true;
                break;
            case 2:
                guac_mouse.currentState.right = true;
                break;
        }
        if (guac_mouse.onmousedown)
            guac_mouse.onmousedown(guac_mouse.currentState);
    }, false);

    element.addEventListener("mouseup", function (e) {
        cancelEvent(e);
        if (ignore_mouse)
            return;
        switch (e.button) {
            case 0:
                guac_mouse.currentState.left = false;
                break;
            case 1:
                guac_mouse.currentState.middle = false;
                break;
            case 2:
                guac_mouse.currentState.right = false;
                break;
        }
        if (guac_mouse.onmouseup)
            guac_mouse.onmouseup(guac_mouse.currentState);
    }, false);

    element.addEventListener("mouseout", function (e) {
        if (!e) e = window.event;
        var target = e.relatedTarget || e.toElement;
        while (target) {
            if (target === element)
                return;
            target = target.parentNode;
        }
        cancelEvent(e);
        if (guac_mouse.currentState.left
            || guac_mouse.currentState.middle
            || guac_mouse.currentState.right) {
            guac_mouse.currentState.left = false;
            guac_mouse.currentState.middle = false;
            guac_mouse.currentState.right = false;
            if (guac_mouse.onmouseup)
                guac_mouse.onmouseup(guac_mouse.currentState);
        }
        if (guac_mouse.onmouseout)
            guac_mouse.onmouseout();
    }, false);

    element.addEventListener("selectstart", function (e) {
        cancelEvent(e);
    }, false);

    function ignorePendingMouseEvents() {
        ignore_mouse = guac_mouse.touchMouseThreshold;
    }

    element.addEventListener("touchmove", ignorePendingMouseEvents, false);
    element.addEventListener("touchstart", ignorePendingMouseEvents, false);
    element.addEventListener("touchend", ignorePendingMouseEvents, false);

    function mousewheel_handler(e) {
        var delta = e.deltaY || -e.wheelDeltaY || -e.wheelDelta;
        if (delta) {
            if (e.deltaMode === 1)
                delta = e.deltaY * guac_mouse.PIXELS_PER_LINE;
            else if (e.deltaMode === 2)
                delta = e.deltaY * guac_mouse.PIXELS_PER_PAGE;
        } else
            delta = e.detail * guac_mouse.PIXELS_PER_LINE;

        scroll_delta += delta;

        if (scroll_delta <= -guac_mouse.scrollThreshold) {
            do {
                if (guac_mouse.onmousedown) {
                    guac_mouse.currentState.up = true;
                    guac_mouse.onmousedown(guac_mouse.currentState);
                }
                if (guac_mouse.onmouseup) {
                    guac_mouse.currentState.up = false;
                    guac_mouse.onmouseup(guac_mouse.currentState);
                }
                scroll_delta += guac_mouse.scrollThreshold;
            } while (scroll_delta <= -guac_mouse.scrollThreshold);
            scroll_delta = 0;
        }

        if (scroll_delta >= guac_mouse.scrollThreshold) {
            do {
                if (guac_mouse.onmousedown) {
                    guac_mouse.currentState.down = true;
                    guac_mouse.onmousedown(guac_mouse.currentState);
                }
                if (guac_mouse.onmouseup) {
                    guac_mouse.currentState.down = false;
                    guac_mouse.onmouseup(guac_mouse.currentState);
                }
                scroll_delta -= guac_mouse.scrollThreshold;
            } while (scroll_delta >= guac_mouse.scrollThreshold);
            scroll_delta = 0;
        }

        cancelEvent(e);
    }

    element.addEventListener('DOMMouseScroll', mousewheel_handler, false);
    element.addEventListener('mousewheel', mousewheel_handler, false);
    element.addEventListener('wheel', mousewheel_handler, false);

    this.setCursor = function (canvas, x, y) {
        if (CSS3_CURSOR_SUPPORTED) {
            var dataURL = canvas.toDataURL('image/png');
            element.style.cursor = "url(" + dataURL + ") " + x + " " + y + ", auto";
            return true;
        }
        return false;
    };

    var CSS3_CURSOR_SUPPORTED = (function () {
        var div = document.createElement("div");
        if (!("cursor" in div.style))
            return false;
        try {
            div.style.cursor = "url(data:image/png;base64,"
                + "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"
                + "AQMAAAAl21bKAAAAA1BMVEX///+nxBvI"
                + "AAAACklEQVQI12NgAAAAAgAB4iG8MwAA"
                + "AABJRU5ErkJggg==) 0 0, auto";
        } catch (e) {
            return false;
        }
        return /\burl\([^()]*\)\s+0\s+0\b/.test(div.style.cursor || "");
    })();
};

mouse.State = Guacamole.Mouse.State
mouse.Touchpad = Guacamole.Mouse.Touchpad
mouse.Touchscreen = Guacamole.Mouse.Touchscreen

export default {
    mouse
}